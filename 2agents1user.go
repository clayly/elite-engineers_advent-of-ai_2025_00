package main

import (
	"context"
)

func Run2Agents1User() error {
	inspector := NewSimpleAgentInspector(nil)
	interviewer := NewAgentInterviewer(nil, inspector)
	return interviewer.Run(context.Background())
}
