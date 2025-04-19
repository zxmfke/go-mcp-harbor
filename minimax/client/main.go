package main

import (
	"context"
	"fmt"
	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"log"
)

func main() {

	//stdio, err := transport.NewStdioClientTransport("./miniMaxMCPServer.exe", nil)
	//if err != nil {
	//	log.Fatalf("new stdio client err : %s", err.Error())
	//}

	sse, err := transport.NewSSEClientTransport("http://127.0.0.1:8080/sse")
	if err != nil {
		log.Fatalf("NewSSEServerTransport failed: %v", err)
	}

	mcpClient, err := client.NewClient(sse, client.WithClientInfo(protocol.Implementation{
		Name:    "minimax mcp client",
		Version: "1.0.0",
	}))
	if err != nil {
		log.Fatalf("NewClient err : %s", err.Error())
	}

	list, err := mcpClient.ListTools(context.Background())
	if err != nil {
		log.Fatalf("ListTools err : %s", err.Error())
	}

	for _, tool := range list.Tools {
		log.Printf("\n%s: %s\n", tool.Name, tool.Description)
	}

	if err := CallTextToImg(mcpClient, "画一只可爱的小蛇，要铅笔画"); err != nil {
		log.Fatalf("CallTextToImg err : %s", err.Error())
	}

	log.Printf("CallTextToImg success")
}

func CallTextToImg(mcpClient *client.Client, prompt string) error {
	resp, err := mcpClient.CallTool(context.Background(), &protocol.CallToolRequest{
		Name: "text_to_image",
		Arguments: map[string]interface{}{
			"prompt": prompt,
		},
	})
	if err != nil {
		return err
	}

	for i := 0; i < len(resp.Content); i++ {
		log.Printf("%+v", resp.Content[i])
	}

	if resp.IsError {
		return fmt.Errorf("call fail")
	}

	return nil
}
