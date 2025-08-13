package main

import (
	"context"
)

// StartAgents constructs an AgentInterviewer with an injected AgentInspector
// and runs it. This keeps main.go unchanged; to use this orchestration,
// call StartAgents from your own entrypoint or tests.
func StartAgents(ctx context.Context) error {
	inspector := &SimpleAgentInspector{}
	interviewer := NewAgentInterviewer(nil, inspector)
	return interviewer.Run(ctx)
}
