package main

import (
	"context"
	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"gopkg.in/ini.v1"
	"log"
)

func main() {
	var (
		sseTransport transport.ClientTransport
		sseClient    *client.Client
		err          error
	)

	// Load configuration from ini file
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	// Read necessary configuration variables
	apiKey := cfg.Section("Gaode").Key("APIKey").String()
	if apiKey == "" {
		log.Fatalf("APIKey not set in config file")
	}

	sseTransport, err = transport.NewSSEClientTransport("https://mcp.amap.com/sse?key=" + apiKey)
	if err != nil {
		log.Fatalf("Failed to create SSE client transport: %v", err)
		return
	}

	sseClient, err = client.NewClient(sseTransport)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return
	}

	_, err = sseClient.ListTools(context.Background())
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
		return
	}

	//for i := 0; i < len(resp.Tools); i++ {
	//	tool := resp.Tools[i]
	//	log.Printf("%+v", tool)
	//}

	result, err := sseClient.CallTool(context.Background(), &protocol.CallToolRequest{
		Name: "maps_geo",
		Arguments: map[string]interface{}{
			"origin":      "厦门软件园三期B16",
			"destination": "厦门高崎机场",
		},
	})

	if err != nil {
		log.Fatalf("Failed to call tool: %v", err)
		return
	}

	if result.IsError {
		log.Printf("call tool fail")
		//return
	}

	for i := 0; i < len(result.Content); i++ {
		log.Printf("%d: %s", i+1, result.Content[i])
	}

}
