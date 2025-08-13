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
}

func NewSimpleAgentInspector(client *openrouter.Client) *SimpleAgentInspector {
	if client == nil {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			log.Fatal("export OPENROUTER_API_KEY first")
		}
		client = openrouter.NewClient(apiKey)
	}
	return &SimpleAgentInspector{client: client}
}

func (s *SimpleAgentInspector) Inspect(ctx context.Context, resp ZRsp) error {
	// Ensure we have a client (constructor should set this; keep a fallback for safety)
	if s.client == nil {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("export OPENROUTER_API_KEY first")
		}
		s.client = openrouter.NewClient(apiKey)
	}

	// Marshal the structured response to JSON
	payload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal ZRsp: %w", err)
	}

	// Prepare a minimal system prompt and send the JSON payload as user content
	sys := "You are an AI inspector. Briefly acknowledge the structured JSON payload and do not include extra prose."
	req := openrouter.ChatCompletionRequest{
		Model: "deepseek/deepseek-chat-v3-0324:free",
		Messages: []openrouter.ChatCompletionMessage{
			{Role: openrouter.ChatMessageRoleSystem, Content: openrouter.Content{Text: sys}},
			{Role: openrouter.ChatMessageRoleUser, Content: openrouter.Content{Text: string(payload)}},
		},
	}

	respLLM, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return fmt.Errorf("inspector openrouter call: %w", err)
	}

	if len(respLLM.Choices) > 0 {
		fmt.Printf("Inspector LLM ack: %s\n", respLLM.Choices[0].Message.Content.Text)
	}
	return nil
}
