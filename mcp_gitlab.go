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

func RunMCPGitLab() {
	gitlabToken := os.Getenv("GITLAB_PERSONAL_ACCESS_TOKEN")
	if gitlabToken == "" {
		log.Fatal("export GITLAB_PERSONAL_ACCESS_TOKEN first")
	}

	// Launch the MCP GitLab server as a subprocess using Docker
	// Assume Docker is installed and the image is pulled: docker pull mcp/gitlab
	// Replace 'your_gitlab_token' with your actual GitLab token
	cmd := exec.Command("docker", "run", "--rm", "-i", "-e", "GITLAB_PERSONAL_ACCESS_TOKEN="+gitlabToken, "mcp/gitlab")

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
	var cursor *string = &str
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

	// Step 2: Call a tool for searching repositories (projects in GitLab)
	// Assuming the tool name is "search_projects" based on common GitLab MCP implementations
	// Adjust the tool name and arguments if different (check from list tools output)
	searchArgs := struct {
		Query string `json:"search"`
	}{
		Query: "vlc", // Replace with your search query
	}

	response, err := client.CallTool(context.Background(), "search_repositories", searchArgs)
	if err != nil {
		log.Fatalf("Failed to call search_repositories tool: %v", err)
	}

	// Print the response
	if response != nil && len(response.Content) > 0 {
		resultJSON, _ := json.MarshalIndent(response.Content[0], "", "  ")
		fmt.Printf("Search Results:\n%s\n", resultJSON)
	} else {
		fmt.Println("No results found.")
	}
}
