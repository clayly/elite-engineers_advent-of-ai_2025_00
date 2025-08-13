package main

import (
	"context"
	"log"
)

func main() {
	if err := StartAgents(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func StartAgents(ctx context.Context) error {
	inspector := NewSimpleAgentInspector(nil)
	interviewer := NewAgentInterviewer(nil, inspector)
	return interviewer.Run(ctx)
}
