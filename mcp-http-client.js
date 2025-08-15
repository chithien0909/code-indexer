#!/usr/bin/env node

/**
 * MCP HTTP Client - Bridges MCP protocol to HTTP API
 * 
 * This client allows VSCode MCP extensions to connect to the 
 * MCP Code Indexer daemon via HTTP API calls.
 * 
 * Usage in VSCode settings.json:
 * {
 *   "mcpServers": {
 *     "CodeIndexer": {
 *       "command": "node",
 *       "args": ["/home/hp/Documents/personal/my-mcp/mcp-http-client.js", "http://localhost:8080"]
 *     }
 *   }
 * }
 */

const http = require('http');
const https = require('https');
const url = require('url');

class MCPHttpClient {
    constructor(baseUrl) {
        this.baseUrl = baseUrl;
        this.sessionId = this.generateSessionId();
        this.tools = [];
        
        // Initialize client
        this.initialize();
    }

    generateSessionId() {
        return `vscode-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    }

    async initialize() {
        try {
            // Get available tools
            const toolsResponse = await this.httpRequest('/api/tools', 'GET');
            this.tools = toolsResponse.tools || [];
            
            // Create session
            await this.createSession();
            
            // Send initialization response
            this.sendResponse({
                jsonrpc: "2.0",
                result: {
                    protocolVersion: "2024-11-05",
                    capabilities: {
                        tools: {
                            listChanged: false
                        }
                    },
                    serverInfo: {
                        name: "MCP Code Indexer HTTP Client",
                        version: "1.0.0"
                    }
                }
            });
        } catch (error) {
            this.sendError(-32603, `Initialization failed: ${error.message}`);
        }
    }

    async createSession() {
        try {
            const sessionData = {
                name: this.sessionId,
                workspace_dir: process.cwd()
            };
            
            await this.httpRequest('/api/sessions', 'POST', sessionData);
        } catch (error) {
            // Session creation failed, but continue anyway
            console.error('Session creation failed:', error.message);
        }
    }

    async httpRequest(path, method = 'GET', data = null) {
        return new Promise((resolve, reject) => {
            const fullUrl = this.baseUrl + path;
            const parsedUrl = url.parse(fullUrl);
            const isHttps = parsedUrl.protocol === 'https:';
            const httpModule = isHttps ? https : http;

            const options = {
                hostname: parsedUrl.hostname,
                port: parsedUrl.port || (isHttps ? 443 : 80),
                path: parsedUrl.path,
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                    'User-Agent': 'MCP-HTTP-Client/1.0.0'
                }
            };

            const req = httpModule.request(options, (res) => {
                let body = '';
                res.on('data', (chunk) => body += chunk);
                res.on('end', () => {
                    try {
                        const response = JSON.parse(body);
                        if (res.statusCode >= 200 && res.statusCode < 300) {
                            resolve(response);
                        } else {
                            reject(new Error(`HTTP ${res.statusCode}: ${response.message || body}`));
                        }
                    } catch (error) {
                        reject(new Error(`Invalid JSON response: ${body}`));
                    }
                });
            });

            req.on('error', reject);

            if (data) {
                req.write(JSON.stringify(data));
            }

            req.end();
        });
    }

    async handleRequest(request) {
        try {
            switch (request.method) {
                case 'tools/list':
                    return this.handleListTools();
                
                case 'tools/call':
                    return await this.handleToolCall(request.params);
                
                default:
                    throw new Error(`Unknown method: ${request.method}`);
            }
        } catch (error) {
            throw error;
        }
    }

    handleListTools() {
        return {
            tools: this.tools.map(tool => ({
                name: tool.name,
                description: tool.description,
                inputSchema: {
                    type: "object",
                    properties: {},
                    required: []
                }
            }))
        };
    }

    async handleToolCall(params) {
        const { name, arguments: args } = params;
        
        try {
            const response = await this.httpRequest('/api/call', 'POST', {
                tool: name,
                arguments: args || {},
                session_id: this.sessionId
            });

            if (response.success) {
                return {
                    content: [
                        {
                            type: "text",
                            text: JSON.stringify(response.result, null, 2)
                        }
                    ]
                };
            } else {
                throw new Error(`Tool execution failed: ${response.error || 'Unknown error'}`);
            }
        } catch (error) {
            throw new Error(`HTTP API call failed: ${error.message}`);
        }
    }

    sendResponse(response) {
        console.log(JSON.stringify(response));
    }

    sendError(code, message, id = null) {
        this.sendResponse({
            jsonrpc: "2.0",
            error: {
                code: code,
                message: message
            },
            id: id
        });
    }

    start() {
        process.stdin.setEncoding('utf8');
        
        let buffer = '';
        process.stdin.on('data', (chunk) => {
            buffer += chunk;
            
            // Process complete JSON messages
            let lines = buffer.split('\n');
            buffer = lines.pop(); // Keep incomplete line in buffer
            
            for (const line of lines) {
                if (line.trim()) {
                    try {
                        const request = JSON.parse(line);
                        this.processRequest(request);
                    } catch (error) {
                        this.sendError(-32700, `Parse error: ${error.message}`);
                    }
                }
            }
        });

        process.stdin.on('end', () => {
            process.exit(0);
        });
    }

    async processRequest(request) {
        try {
            const result = await this.handleRequest(request);
            this.sendResponse({
                jsonrpc: "2.0",
                result: result,
                id: request.id
            });
        } catch (error) {
            this.sendError(-32603, error.message, request.id);
        }
    }
}

// Main execution
if (require.main === module) {
    const baseUrl = process.argv[2] || 'http://localhost:8080';
    const client = new MCPHttpClient(baseUrl);
    client.start();
}

module.exports = MCPHttpClient;
