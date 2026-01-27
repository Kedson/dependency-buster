/**
 * Authentication Layer for MCP
 * Supports static tokens (like Backstage) and basic API key auth
 */

import { AuthenticationError } from './errors.js';

export interface AuthConfig {
  /**
   * Enable/disable authentication
   */
  enabled: boolean;
  
  /**
   * Static tokens for authentication
   */
  staticTokens?: string[];
  
  /**
   * Optional: Environment variable name for token
   */
  tokenEnvVar?: string;
  
  /**
   * Allow unauthenticated access to these methods
   */
  publicMethods?: string[];
}

export interface Credentials {
  /**
   * Type of authentication used
   */
  type: 'static_token' | 'api_key' | 'anonymous';
  
  /**
   * Subject/identity (e.g., 'mcp-client', 'user-123')
   */
  subject?: string;
  
  /**
   * Token that was used (for audit logging)
   */
  tokenHash?: string;
  
  /**
   * Additional context passed with credentials
   */
  context?: Record<string, unknown>;
}

export interface RequestContext {
  /**
   * Authenticated credentials
   */
  credentials: Credentials;
  
  /**
   * Request metadata
   */
  requestId: string;
  timestamp: Date;
  
  /**
   * Client information
   */
  clientInfo?: {
    name?: string;
    version?: string;
  };
}

/**
 * Default configuration
 */
const defaultConfig: AuthConfig = {
  enabled: false,
  tokenEnvVar: 'MCP_TOKEN',
  publicMethods: ['initialize', 'tools/list'],
};

let authConfig: AuthConfig = { ...defaultConfig };

/**
 * Configure authentication
 */
export function configureAuth(config: Partial<AuthConfig>): void {
  authConfig = { ...defaultConfig, ...config };
  
  // Load token from environment if specified
  if (authConfig.tokenEnvVar) {
    const envToken = process.env[authConfig.tokenEnvVar];
    if (envToken) {
      authConfig.staticTokens = authConfig.staticTokens || [];
      if (!authConfig.staticTokens.includes(envToken)) {
        authConfig.staticTokens.push(envToken);
      }
    }
  }
}

/**
 * Hash a token for logging (don't log actual tokens)
 */
function hashToken(token: string): string {
  const crypto = require('crypto');
  return crypto.createHash('sha256').update(token).digest('hex').slice(0, 16);
}

/**
 * Validate authentication for a request
 */
export function validateAuth(
  method: string,
  headers?: Record<string, string>
): Credentials {
  // If auth is disabled, return anonymous credentials
  if (!authConfig.enabled) {
    return { type: 'anonymous' };
  }
  
  // Check if this is a public method
  if (authConfig.publicMethods?.includes(method)) {
    return { type: 'anonymous' };
  }
  
  // Extract token from headers
  const authHeader = headers?.['authorization'] || headers?.['Authorization'];
  if (!authHeader) {
    throw new AuthenticationError('Authentication required: No authorization header');
  }
  
  // Support "Bearer <token>" format
  let token: string;
  if (authHeader.startsWith('Bearer ')) {
    token = authHeader.slice(7);
  } else {
    token = authHeader;
  }
  
  // Validate against static tokens
  if (authConfig.staticTokens?.includes(token)) {
    return {
      type: 'static_token',
      subject: 'mcp-client',
      tokenHash: hashToken(token),
    };
  }
  
  throw new AuthenticationError('Invalid token');
}

/**
 * Create a request context
 */
export function createRequestContext(
  credentials: Credentials,
  clientInfo?: { name?: string; version?: string }
): RequestContext {
  return {
    credentials,
    requestId: generateRequestId(),
    timestamp: new Date(),
    clientInfo,
  };
}

/**
 * Generate a unique request ID
 */
function generateRequestId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).slice(2, 8);
  return `req_${timestamp}_${random}`;
}

/**
 * Check if credentials allow a specific action
 */
export function authorizeAction(
  credentials: Credentials,
  action: string,
  resource?: string
): boolean {
  // For now, any authenticated user can perform any action
  // This can be extended with role-based access control
  if (credentials.type === 'anonymous') {
    return false;
  }
  
  return true;
}

/**
 * Middleware for HTTP server authentication
 */
export function authMiddleware(
  req: { headers: Record<string, string>; method: string; url: string },
  method: string
): RequestContext {
  const credentials = validateAuth(method, req.headers);
  return createRequestContext(credentials);
}

/**
 * Generate a secure random token (for setup)
 */
export function generateToken(): string {
  const crypto = require('crypto');
  return crypto.randomBytes(24).toString('base64');
}

/**
 * Get auth configuration (for debugging/info)
 */
export function getAuthInfo(): { enabled: boolean; methods: string[] } {
  return {
    enabled: authConfig.enabled,
    methods: ['static_token'],
  };
}
