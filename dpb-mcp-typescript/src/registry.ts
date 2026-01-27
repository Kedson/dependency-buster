/**
 * Dynamic Action Registry for MCP
 * Allows runtime registration and discovery of actions
 */

import { ToolAnnotations, getToolAnnotations } from './annotations.js';
import { RequestContext } from './auth.js';
import { ValidationError, NotFoundError } from './errors.js';
import { z, ZodSchema } from 'zod';

export interface ActionSchema<TInput = unknown, TOutput = unknown> {
  input: ZodSchema<TInput>;
  output: ZodSchema<TOutput>;
}

export interface ActionDefinition<TInput = unknown, TOutput = unknown> {
  /**
   * Unique action identifier
   */
  name: string;
  
  /**
   * Human-readable title
   */
  title: string;
  
  /**
   * Description of what this action does
   */
  description: string;
  
  /**
   * Zod schemas for input/output validation
   */
  schema: ActionSchema<TInput, TOutput>;
  
  /**
   * Tool annotations
   */
  annotations?: ToolAnnotations;
  
  /**
   * The action handler
   */
  handler: (input: TInput, context: RequestContext) => Promise<TOutput>;
  
  /**
   * Plugin that registered this action
   */
  pluginId?: string;
}

export interface RegisteredAction<TInput = unknown, TOutput = unknown> extends ActionDefinition<TInput, TOutput> {
  /**
   * Unique ID assigned by registry
   */
  id: string;
  
  /**
   * When the action was registered
   */
  registeredAt: Date;
}

/**
 * Convert Zod schema to JSON Schema (simplified)
 */
function zodToJsonSchema(schema: ZodSchema): Record<string, unknown> {
  // This is a simplified conversion - full conversion would use zod-to-json-schema
  const description = schema.description;
  
  // Get the type info
  const def = (schema as any)._def;
  const typeName = def?.typeName;
  
  switch (typeName) {
    case 'ZodString':
      return { type: 'string', description };
    case 'ZodNumber':
      return { type: 'number', description };
    case 'ZodBoolean':
      return { type: 'boolean', description };
    case 'ZodObject':
      const shape = def.shape();
      const properties: Record<string, unknown> = {};
      const required: string[] = [];
      
      for (const [key, value] of Object.entries(shape)) {
        properties[key] = zodToJsonSchema(value as ZodSchema);
        // Check if required (not optional)
        if (!(value as any).isOptional?.()) {
          required.push(key);
        }
      }
      
      return { type: 'object', properties, required, description };
    default:
      return { type: 'string', description };
  }
}

/**
 * Dynamic Action Registry
 */
export class ActionRegistry {
  private actions = new Map<string, RegisteredAction<unknown, unknown>>();
  private actionCounter = 0;
  
  /**
   * Register a new action
   */
  register<TInput, TOutput>(definition: ActionDefinition<TInput, TOutput>): string {
    const id = `action_${++this.actionCounter}`;
    
    // Validate that action name is unique
    if (this.actions.has(definition.name)) {
      throw new ValidationError(`Action "${definition.name}" is already registered`);
    }
    
    const registered: RegisteredAction<TInput, TOutput> = {
      ...definition,
      id,
      registeredAt: new Date(),
      annotations: definition.annotations || getToolAnnotations(definition.name),
    };
    
    this.actions.set(definition.name, registered as RegisteredAction<unknown, unknown>);
    console.error(`[Registry] Registered action: ${definition.name} (${id})`);
    
    return id;
  }
  
  /**
   * Unregister an action
   */
  unregister(name: string): boolean {
    const deleted = this.actions.delete(name);
    if (deleted) {
      console.error(`[Registry] Unregistered action: ${name}`);
    }
    return deleted;
  }
  
  /**
   * Get action by name
   */
  get(name: string): RegisteredAction | undefined {
    return this.actions.get(name);
  }
  
  /**
   * List all registered actions
   */
  list(options?: { pluginId?: string }): RegisteredAction[] {
    let actions = Array.from(this.actions.values());
    
    if (options?.pluginId) {
      actions = actions.filter(a => a.pluginId === options.pluginId);
    }
    
    return actions;
  }
  
  /**
   * Invoke an action
   */
  async invoke(
    name: string,
    input: unknown,
    context: RequestContext
  ): Promise<unknown> {
    const action = this.actions.get(name);
    
    if (!action) {
      throw new NotFoundError(`Action "${name}" not found`);
    }
    
    // Validate input
    const parseResult = action.schema.input.safeParse(input);
    if (!parseResult.success) {
      throw new ValidationError(
        `Invalid input for action "${name}": ${parseResult.error.message}`,
        parseResult.error.issues
      );
    }
    
    // Execute the handler
    const output = await action.handler(parseResult.data, context);
    
    // Validate output (optional, for debugging)
    const outputResult = action.schema.output.safeParse(output);
    if (!outputResult.success) {
      console.error(`[Registry] Warning: Output validation failed for ${name}`);
    }
    
    return output;
  }
  
  /**
   * Convert actions to MCP tool format
   */
  toMcpTools(): Array<{
    name: string;
    description: string;
    inputSchema: Record<string, unknown>;
    annotations?: ToolAnnotations;
  }> {
    return this.list().map(action => ({
      name: action.name,
      description: action.description,
      inputSchema: zodToJsonSchema(action.schema.input),
      annotations: action.annotations,
    }));
  }
}

/**
 * Global action registry instance
 */
export const registry = new ActionRegistry();

/**
 * Convenience function to register an action
 */
export function registerAction<TInput, TOutput>(
  definition: ActionDefinition<TInput, TOutput>
): string {
  return registry.register(definition);
}

/**
 * Convenience decorator-style function for registering actions
 */
export function action<TInput, TOutput>(
  name: string,
  title: string,
  description: string,
  inputSchema: ZodSchema<TInput>,
  outputSchema: ZodSchema<TOutput>,
  options?: { annotations?: ToolAnnotations; pluginId?: string }
) {
  return (handler: (input: TInput, context: RequestContext) => Promise<TOutput>) => {
    return registerAction({
      name,
      title,
      description,
      schema: { input: inputSchema, output: outputSchema },
      handler,
      ...options,
    });
  };
}
