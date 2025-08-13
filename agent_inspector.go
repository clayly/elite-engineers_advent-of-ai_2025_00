package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/revrost/go-openrouter"
)

type AgentInspector interface {
	Inspect(ctx context.Context, resp ZRsp) error
}

type SimpleAgentInspector struct {
	client *openrouter.Client

	sysPrompt string
}

func NewSimpleAgentInspector(client *openrouter.Client) *SimpleAgentInspector {
	if client == nil {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			log.Fatal("export OPENROUTER_API_KEY first")
		}
		client = openrouter.NewClient(apiKey)
	}

	agent := &SimpleAgentInspector{
		client: client,
	}
	agent.sysPrompt = "You are an AI inspector. Briefly acknowledge the structured JSON payload and do not include extra prose."
	return agent
}

func (agent *SimpleAgentInspector) Inspect(ctx context.Context, zResp ZRsp) error {
	zRespStr, err := json.Marshal(zResp)
	if err != nil {
		return fmt.Errorf("marshal ZRsp: %w", err)
	}

	resp, err := agent.client.CreateChatCompletion(
		ctx,
		openrouter.ChatCompletionRequest{
			Model: "deepseek/deepseek-chat-v3-0324:free",
			Messages: []openrouter.ChatCompletionMessage{
				{Role: openrouter.ChatMessageRoleSystem, Content: openrouter.Content{Text: agent.sysPrompt}},
				{Role: openrouter.ChatMessageRoleUser, Content: openrouter.Content{Text: string(zRespStr)}},
			},
		},
	)

	if err != nil {
		return fmt.Errorf("inspector openrouter call: %w", err)
	}

	if len(resp.Choices) > 0 {
		fmt.Printf("Inspector LLM ack: %s\n", resp.Choices[0].Message.Content.Text)
	}
	return nil
}
