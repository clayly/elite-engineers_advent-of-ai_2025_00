package main

import (
	"context"
	"fmt"
)

// AgentInspector is responsible for inspecting structured ZRsp results
// produced by the AgentInterviewer before the interviewer proceeds.
type AgentInspector interface {
	Inspect(ctx context.Context, resp ZRsp) error
}

// SimpleAgentInspector is a minimal implementation that prints
// the structured response and returns immediately.
type SimpleAgentInspector struct{}

func (s *SimpleAgentInspector) Inspect(ctx context.Context, resp ZRsp) error {
	// In a real implementation, this could validate schema, persist,
	// or trigger side-effects. Here we just log briefly.
	fmt.Printf("Inspector: received structured response with %d item(s)\n", len(resp.Items))
	for i, it := range resp.Items {
		fmt.Printf("  #%d: type=%s name=%s %s=%s %s\n", i+1, it.ItemType, it.ItemName, it.Value1Name, it.Value1, it.Value1Units)
	}
	return nil
}
