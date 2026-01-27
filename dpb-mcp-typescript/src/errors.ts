/**
 * Typed Error System for MCP
 * Based on Backstage error patterns
 */

// Standard MCP error codes (JSON-RPC compatible)
export const ErrorCodes = {
  // Standard JSON-RPC errors
  PARSE_ERROR: -32700,
  INVALID_REQUEST: -32600,
  METHOD_NOT_FOUND: -32601,
  INVALID_PARAMS: -32602,
  INTERNAL_ERROR: -32603,
  
  // Custom MCP errors
  NOT_FOUND: -32000,
  NOT_ALLOWED: -32001,
  VALIDATION_ERROR: -32002,
  AUTHENTICATION_ERROR: -32003,
  RATE_LIMITED: -32004,
  TIMEOUT: -32005,
} as const;

export class McpError extends Error {
  readonly code: number;
  readonly data?: unknown;

  constructor(message: string, code: number, data?: unknown) {
    super(message);
    this.name = 'McpError';
    this.code = code;
    this.data = data;
    Object.setPrototypeOf(this, McpError.prototype);
  }

  toJSON() {
    return {
      code: this.code,
      message: this.message,
      data: this.data,
    };
  }
}

/**
 * Resource not found error
 * Use when: requested resource (file, repo, package) doesn't exist
 */
export class NotFoundError extends McpError {
  constructor(message: string, data?: unknown) {
    super(message, ErrorCodes.NOT_FOUND, data);
    this.name = 'NotFoundError';
  }
}

/**
 * Permission denied error
 * Use when: user lacks required permissions for the operation
 */
export class NotAllowedError extends McpError {
  constructor(message: string, data?: unknown) {
    super(message, ErrorCodes.NOT_ALLOWED, data);
    this.name = 'NotAllowedError';
  }
}

/**
 * Validation error
 * Use when: input parameters fail validation
 */
export class ValidationError extends McpError {
  constructor(message: string, data?: unknown) {
    super(message, ErrorCodes.VALIDATION_ERROR, data);
    this.name = 'ValidationError';
  }
}

/**
 * Authentication error
 * Use when: authentication is required but failed or missing
 */
export class AuthenticationError extends McpError {
  constructor(message: string, data?: unknown) {
    super(message, ErrorCodes.AUTHENTICATION_ERROR, data);
    this.name = 'AuthenticationError';
  }
}

/**
 * Timeout error
 * Use when: operation exceeded time limit
 */
export class TimeoutError extends McpError {
  constructor(message: string, data?: unknown) {
    super(message, ErrorCodes.TIMEOUT, data);
    this.name = 'TimeoutError';
  }
}

/**
 * Convert any error to a structured MCP error response
 */
export function toMcpError(error: unknown): McpError {
  if (error instanceof McpError) {
    return error;
  }
  
  if (error instanceof Error) {
    // Check for common error patterns
    const msg = error.message.toLowerCase();
    
    if (msg.includes('enoent') || msg.includes('not found') || msg.includes('does not exist')) {
      return new NotFoundError(error.message);
    }
    
    if (msg.includes('permission denied') || msg.includes('access denied') || msg.includes('forbidden')) {
      return new NotAllowedError(error.message);
    }
    
    if (msg.includes('invalid') || msg.includes('required')) {
      return new ValidationError(error.message);
    }
    
    if (msg.includes('unauthorized') || msg.includes('authentication')) {
      return new AuthenticationError(error.message);
    }
    
    if (msg.includes('timeout') || msg.includes('timed out')) {
      return new TimeoutError(error.message);
    }
    
    return new McpError(error.message, ErrorCodes.INTERNAL_ERROR, {
      stack: error.stack,
    });
  }
  
  return new McpError(String(error), ErrorCodes.INTERNAL_ERROR);
}
