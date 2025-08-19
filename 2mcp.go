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

func Run2MCP() {
	githubToken := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if githubToken == "" {
		log.Fatal("export GITHUB_PERSONAL_ACCESS_TOKEN first")
	}

	rsp1, err := cmdForResult(
		exec.Command(
			"docker",
			"run",
			"--rm",
			"-i",
			"-e",
			"GITHUB_PERSONAL_ACCESS_TOKEN="+githubToken,
			"ghcr.io/github/github-mcp-server",
		),
		"list_notifications",
		struct{}{},
	)
	if err != nil {
		log.Fatalf("cmd1 Failed cmdForResult: %v", err)
	}

	var rsp1Str = ""
	if rsp1 != nil && len(rsp1.Content) > 0 {
		mcpRspBytes, _ := json.MarshalIndent(rsp1.Content[0], "", "  ")
		rsp1Str = string(mcpRspBytes)
		fmt.Printf("cmd1 tool results:\n%s\n", rsp1Str)
	} else {
		fmt.Println("cmd1 no tool results")
	}

	rsp2, err := cmdForResult(
		exec.Command(
			"docker",
			"run",
			"-i",
			"--rm",
			"--mount",
			"type=bind,src=/home/clayly/projects/course/elite-engineers_advent-of-ai_2025_00/tmp,dst=/projects",
			"mcp/filesystem",
			"/projects",
		),
		"write_file",
		struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}{
			Path:    "/projects/test.txt", // Path inside the container's allowed directory
			Content: rsp1Str,
		},
	)
	if err != nil {
		log.Fatalf("cmd1 Failed cmdForResult: %v", err)
	}

	if rsp2 != nil && len(rsp2.Content) > 0 {
		rspBytes, _ := json.MarshalIndent(rsp2.Content[0], "", "  ")
		rspStr := string(rspBytes)
		fmt.Printf("cmd2 tool results:\n%s\n", rspStr)
	} else {
		fmt.Println("cmd2 no tool results")
	}
}

func cmdForResult(cmd *exec.Cmd, toolName string, toolArguments any) (*mcp.ToolResponse, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("%s Failed to get stdin pipe: %v", toolName, err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("%s Failed to get stdout pipe: %v", toolName, err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("%s Failed to start docker container: %v", toolName, err)
	}
	defer func(Process *os.Process) {
		err := Process.Kill()
		if err != nil {
			log.Printf("%s Warning: Failed to kill docker container: %v", toolName, err)
		}
	}(cmd.Process)

	transport := stdio.NewStdioServerTransportWithIO(stdout, stdin)
	client := mcp.NewClient(transport)

	if _, err := client.Initialize(context.Background()); err != nil {
		log.Printf("%s Warning: Failed to initialize: %v", toolName, err)
	}

	str := ""
	var cursor = &str
	for {
		tools, err := client.ListTools(context.Background(), cursor)
		if err != nil {
			log.Fatalf("%s Failed to list tools: %v", toolName, err)
		}

		if tools.NextCursor == nil {
			break
		}
		cursor = tools.NextCursor
	}

	return client.CallTool(context.Background(), toolName, toolArguments)
}
