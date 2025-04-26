package main

import (
	"log"
	"mcp/minimax/server/define"
	"mcp/minimax/server/minimax"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"gopkg.in/ini.v1"
)

func main() {
	// Load configuration from ini file
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	// Read necessary configuration variables
	apiKey := cfg.Section("Minimax").Key("APIKey").String()
	if apiKey == "" {
		log.Fatalf("APIKey not set in config file")
	}

	apiHost := cfg.Section("Minimax").Key("APIHost").String()
	if apiHost == "" {
		log.Fatalf("APIHost not set in config file")
	}

	mode := define.ServerMode(cfg.Section("Minimax").Key("Mode").String())
	if !mode.Valid() {
		log.Fatalf("Mode not set in config file")
	}
	addr := cfg.Section("Minimax").Key("Addr").String()
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	resourceMode := cfg.Section("Minimax").Key("ResourceMode").String()
	if resourceMode == "" {
		resourceMode = "url" // Default value
	}

	// Create Minimax API client
	apiClient := &minimax.APIClient{
		APIKey:  apiKey,
		APIHost: apiHost,
	}

	apiServer := &minimax.MCPServer{
		Client:       apiClient,
		ResourceMode: resourceMode,
	}

	// Create MCP transport layer
	var transportServer transport.ServerTransport
	switch mode {
	case define.SSE:
		transportServer, err = transport.NewSSEServerTransport(addr)
		if err != nil {
			log.Fatalf("Failed to create SSE transport: %v", err)
		}
	case define.Streamable:
		transportServer = transport.NewStreamableHTTPServerTransport(addr,
			transport.WithStreamableHTTPServerTransportOptionStateMode(transport.Stateful))
	default:
		transportServer = transport.NewStdioServerTransport()
	}

	// Create MCP server
	mcpServer, err := server.NewServer(transportServer,
		server.WithServerInfo(protocol.Implementation{
			Name:    "MiniMax MCP",
			Version: "1.0.0",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Register tools
	minimax.RegisterTools(mcpServer, apiServer)

	if err = mcpServer.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
