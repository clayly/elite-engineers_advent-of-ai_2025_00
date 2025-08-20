package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

func RunMCPGithub() {
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
	client := mcp.NewClient(transport)

	// Initialize the client (may not be required for stdio, but good practice)
	if _, err := client.Initialize(context.Background()); err != nil {
		log.Printf("Warning: Failed to initialize client: %v", err)
	}

	// Step 1: List available tools
	str := ""
	var cursor = &str
	for {
		tools, err := client.ListTools(context.Background(), cursor)
		if err != nil {
			log.Fatalf("Failed to list tools: %v", err)
		}

		// Process tools...

		if tools.NextCursor == nil {
			break // No more pages
		}
		cursor = tools.NextCursor
	}

	toolName := "list_notifications"
	toolArgs := struct{}{}
	response, err := client.CallTool(context.Background(), toolName, toolArgs)

	if err != nil {
		log.Fatalf("failed to call tool: %v", err)
	}

	// Print the response
	if response != nil && len(response.Content) > 0 {
		resultJSON, _ := json.MarshalIndent(response.Content[0], "", "  ")
		fmt.Printf("tool results:\n%s\n", resultJSON)
	} else {
		fmt.Println("no tool results")
	}
}
