package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/revrost/go-openrouter"
)

//  There is my github notifications from guthub below. Print some short summary of it.

func RunMCPGithubAndLlm() {
	githubToken := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if githubToken == "" {
		log.Fatal("export GITHUB_PERSONAL_ACCESS_TOKEN first")
	}

	// Launch the MCP GitLab server as a subprocess using Docker
	// Assume Docker is installed and the image is pulled: docker pull mcp/gitlab
	// Replace 'your_gitlab_token' with your actual GitLab token
	cmd := exec.Command("docker", "run", "--rm", "-i", "-e", "GITHUB_PERSONAL_ACCESS_TOKEN="+githubToken, "ghcr.io/github/github-mcp-server")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start docker container: %v", err)
	}
	defer func(Process *os.Process) {
		err := Process.Kill()
		if err != nil {
			log.Printf("Warning: Failed to kill docker container: %v", err)
		}
	}(cmd.Process)

	// Create stdio transport for the MCP client
	transport := stdio.NewStdioServerTransportWithIO(stdout, stdin)

	// Create the MCP client
	mcpClient := mcp.NewClient(transport)

	// Initialize the client (may not be required for stdio, but good practice)
	if _, err := mcpClient.Initialize(context.Background()); err != nil {
		log.Printf("Warning: Failed to initialize client: %v", err)
	}

	// Step 1: List available tools
	str := ""
	var cursor = &str
	for {
		tools, err := mcpClient.ListTools(context.Background(), cursor)
		if err != nil {
			log.Fatalf("Failed to list tools: %v", err)
		}

		// Process tools...

		if tools.NextCursor == nil {
			break // No more pages
		}
		cursor = tools.NextCursor
	}

	//toolName := "search_repositories"
	//toolArgs := struct {
	//	Query string `json:"query"`
	//}{
	//	Query: "test-archiver", // Replace with your search query
	//}
	//response, err := client.CallTool(context.Background(), toolName, toolArgs)

	toolName := "list_notifications"
	toolArgs := struct{}{}
	mcpRsp, err := mcpClient.CallTool(context.Background(), toolName, toolArgs)

	if err != nil {
		log.Fatalf("failed to call tool: %v", err)
	}

	// Print the response
	var mspRspStr = ""
	if mcpRsp != nil && len(mcpRsp.Content) > 0 {
		mcpRspBytes, _ := json.MarshalIndent(mcpRsp.Content[0], "", "  ")
		mspRspStr = string(mcpRspBytes)
		fmt.Printf("tool results:\n%s\n", mspRspStr)
	} else {
		fmt.Println("no tool results")
	}

	llmToken := os.Getenv("OPENROUTER_API_KEY")
	if llmToken == "" {
		log.Fatal("export OPENROUTER_API_KEY first")
	}
	llmClient := openrouter.NewClient(llmToken)

	fmt.Print("\nПешы: ")
	reader := bufio.NewReader(os.Stdin)
	userInput, _ := reader.ReadString('\n')

	llmReqStr := strings.TrimSpace(userInput) + "\n" + mspRspStr

	llmReqStrEscaped, err := json.Marshal(llmReqStr)
	if err != nil {
		log.Fatalf("failed to read user input: %v", err)
	}

	resp, err := llmClient.CreateChatCompletion(
		context.Background(),
		openrouter.ChatCompletionRequest{
			Model: "deepseek/deepseek-chat-v3-0324:free",
			Messages: []openrouter.ChatCompletionMessage{
				{Role: openrouter.ChatMessageRoleUser, Content: openrouter.Content{Text: string(llmReqStrEscaped)}},
			},
		},
	)

	if err != nil {
		log.Fatalf("llm err: %v", err)
	}

	fmt.Printf("llm rsp: %s\n", resp.Choices[0].Message.Content.Text)
}
