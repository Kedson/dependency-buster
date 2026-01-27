/**
 * HTTP/SSE Transport for MCP
 * Provides HTTP endpoints alongside stdio transport
 */

import { createServer, IncomingMessage, ServerResponse } from 'http';
import { validateAuth, createRequestContext, RequestContext } from './auth.js';
import { McpError, ErrorCodes, toMcpError } from './errors.js';

export interface HttpConfig {
  port: number;
  host?: string;
  basePath?: string;
  corsOrigins?: string[];
}

export interface SseClient {
  id: string;
  response: ServerResponse;
  context: RequestContext;
}

export type JsonRpcHandler = (
  method: string,
  params: Record<string, unknown>,
  context: RequestContext
) => Promise<unknown>;

const defaultConfig: HttpConfig = {
  port: 3000,
  host: '127.0.0.1',
  basePath: '/api/mcp',
  corsOrigins: ['*'],
};

/**
 * Create HTTP/SSE transport
 */
export function createHttpTransport(
  handler: JsonRpcHandler,
  config: Partial<HttpConfig> = {}
) {
  const cfg = { ...defaultConfig, ...config };
  const sseClients = new Map<string, SseClient>();
  let clientIdCounter = 0;

  const server = createServer((req, res) => {
    // CORS headers
    const origin = req.headers.origin || '*';
    if (cfg.corsOrigins?.includes('*') || cfg.corsOrigins?.includes(origin)) {
      res.setHeader('Access-Control-Allow-Origin', origin);
    }
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');
    
    // Handle preflight
    if (req.method === 'OPTIONS') {
      res.writeHead(204);
      res.end();
      return;
    }

    const url = new URL(req.url || '/', `http://${req.headers.host}`);
    const path = url.pathname;

    // Route handling
    if (path === `${cfg.basePath}/v1` || path === `${cfg.basePath}/v1/`) {
      handleStreamableHttp(req, res, handler);
    } else if (path === `${cfg.basePath}/v1/sse`) {
      handleSse(req, res, sseClients, handler);
    } else if (path === `${cfg.basePath}/health`) {
      handleHealth(res);
    } else if (path === `${cfg.basePath}/info`) {
      handleInfo(res);
    } else {
      res.writeHead(404, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'Not found' }));
    }
  });

  return {
    start: () => {
      server.listen(cfg.port, cfg.host, () => {
        console.error(`HTTP/SSE transport listening on http://${cfg.host}:${cfg.port}${cfg.basePath}/v1`);
        console.error(`SSE endpoint: http://${cfg.host}:${cfg.port}${cfg.basePath}/v1/sse`);
      });
    },
    stop: () => {
      sseClients.forEach(client => client.response.end());
      sseClients.clear();
      server.close();
    },
    broadcast: (event: string, data: unknown) => {
      sseClients.forEach(client => {
        sendSseEvent(client.response, event, data);
      });
    },
  };

  function generateClientId(): string {
    return `client_${++clientIdCounter}_${Date.now().toString(36)}`;
  }

  /**
   * Handle Streamable HTTP (JSON-RPC over HTTP POST)
   */
  async function handleStreamableHttp(
    req: IncomingMessage,
    res: ServerResponse,
    handler: JsonRpcHandler
  ) {
    if (req.method !== 'POST') {
      res.writeHead(405, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'Method not allowed' }));
      return;
    }

    let body = '';
    req.on('data', chunk => body += chunk);
    req.on('end', async () => {
      try {
        const request = JSON.parse(body);
        
        // Validate authentication
        const headers = req.headers as Record<string, string>;
        const credentials = validateAuth(request.method, headers);
        const context = createRequestContext(credentials);
        
        // Handle JSON-RPC request
        const result = await handler(request.method, request.params || {}, context);
        
        const response = {
          jsonrpc: '2.0',
          id: request.id,
          result,
        };
        
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify(response));
      } catch (error) {
        const mcpError = toMcpError(error);
        
        const response = {
          jsonrpc: '2.0',
          id: null,
          error: mcpError.toJSON(),
        };
        
        const statusCode = mcpError.code === ErrorCodes.AUTHENTICATION_ERROR ? 401 : 500;
        res.writeHead(statusCode, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify(response));
      }
    });
  }

  /**
   * Handle SSE connection
   */
  function handleSse(
    req: IncomingMessage,
    res: ServerResponse,
    clients: Map<string, SseClient>,
    handler: JsonRpcHandler
  ) {
    // Validate authentication
    const headers = req.headers as Record<string, string>;
    let context: RequestContext;
    try {
      const credentials = validateAuth('sse/connect', headers);
      context = createRequestContext(credentials);
    } catch (error) {
      res.writeHead(401, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'Unauthorized' }));
      return;
    }

    // Set SSE headers
    res.writeHead(200, {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection': 'keep-alive',
    });

    const clientId = generateClientId();
    clients.set(clientId, { id: clientId, response: res, context });

    // Send connected event
    sendSseEvent(res, 'connected', { clientId });

    // Keep alive
    const keepAlive = setInterval(() => {
      res.write(': keep-alive\n\n');
    }, 30000);

    // Handle disconnect
    req.on('close', () => {
      clearInterval(keepAlive);
      clients.delete(clientId);
    });
  }

  function sendSseEvent(res: ServerResponse, event: string, data: unknown) {
    res.write(`event: ${event}\n`);
    res.write(`data: ${JSON.stringify(data)}\n\n`);
  }

  function handleHealth(res: ServerResponse) {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ status: 'healthy', timestamp: new Date().toISOString() }));
  }

  function handleInfo(res: ServerResponse) {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({
      name: 'php-dependency-analyzer',
      version: '1.0.0',
      protocols: ['stdio', 'http', 'sse'],
      endpoints: {
        http: `${cfg.basePath}/v1`,
        sse: `${cfg.basePath}/v1/sse`,
        health: `${cfg.basePath}/health`,
      },
    }));
  }
}
