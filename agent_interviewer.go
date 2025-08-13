package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/revrost/go-openrouter"
)

type AgentInterviewer struct {
	client    *openrouter.Client
	reader    *bufio.Reader
	inspector AgentInspector

	zProvideDataStart string
	zProvideDataEnd   string
	zCollectDataStart string
	zCollectDataEnd   string
	zRspStart         string
	zRspEnd           string
	zRspFormat        string
	zRspFormatPrompt  string
	sysPrompt         string
	basicPrompt       string
	zDialog           string
}

func NewAgentInterviewer(client *openrouter.Client, inspector AgentInspector) *AgentInterviewer {
	if client == nil {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			log.Fatal("export OPENROUTER_API_KEY first")
		}
		client = openrouter.NewClient(apiKey)
	}
	if inspector == nil {
		inspector = NewSimpleAgentInspector(client)
	}

	agent := &AgentInterviewer{
		client:            client,
		reader:            bufio.NewReader(os.Stdin),
		inspector:         inspector,
		zProvideDataStart: "Z_PROVIDE_DATA_START",
		zProvideDataEnd:   "Z_PROVIDE_DATA_END",
		zCollectDataStart: "Z_COLLECT_DATA_START",
		zCollectDataEnd:   "Z_COLLECT_DATA_END",
		zRspStart:         "Z_RSP_START",
		zRspEnd:           "Z_RSP_END",
		zRspFormat:        "JSON",
	}

	// Build prompts and template exactly like in 1agent1user.go
	agent.zRspFormatPrompt = fmt.Sprintf("exact string %s, right after that valid %s, right after that exact string %s", agent.zRspStart, agent.zRspFormat, agent.zRspEnd)
	fmt.Printf("zRspFormatPrompt=%s\n", agent.zRspFormatPrompt)

	zRspTemplate := ZRsp{
		Items: []ZRspItem{
			{ItemType: "car", ItemName: "TT-34", Value1Name: "cost", Value1Units: "byn", Value1: "1000000"},
			{ItemType: "bullet", ItemName: "7.62x39mm", Value1Name: "speed", Value1Units: "km/h", Value1: "360"},
			{ItemType: "action", ItemName: "deleting folder in Linux", Value1Name: "bash command", Value1Units: "bash code", Value1: "sudo rm -rf {folder_name}"},
		},
	}
	zRspTemplateStr, err := json.Marshal(zRspTemplate)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("zRspTemplateStr=%s\n", string(zRspTemplateStr))

	agent.basicPrompt = fmt.Sprintf(`
There is dialog (named Z_DIALOG) between me (the user, named Z_USER) and you (the AI, named Z_AI).
Z_DIALOG starts after word Z_DIALOG_START.

Z_AI can only do 2 things in Z_DIALOG:
1. Z_AI can give an answer (named Z_RSP, format of which is described below), which will finish Z_DIALOG.
2. Z_AI can collect data from Z_USER with clarifying questions (named Z_COLLECT_DATA) in order to qualitatively fill out Z_RSP with specific information.

Z_AI main and only goal in Z_DIALOG is to collect enough specific data to fill Z_RSP.
Z_AI most ensure that it is collected enough specific data to fill Z_RSP.

Z_USER can only do 2 things in Z_DIALOG:
1. Z_USER can provide additional data with clarifying answers (named Z_PROVIDE_DATA). 
2. Z_USER determines direction of Z_DIALOG and therefore content of Z_RSP with his first Z_PROVIDE_DATA.

All Z_COLLECT_DATA in Z_DIALOG placed between words Z_COLLECT_DATA_START and Z_COLLECT_DATA_END.

All Z_PROVIDE_DATA in Z_DIALOG placed between words Z_PROVIDE_DATA_START and Z_PROVIDE_DATA_END.

Z_AI finishes Z_DIALOG with Z_RSP in 2 occasions:
1. When Z_USER when the user clearly writes that he can't provide more data.
2. When Z_USER is stopped providing relative data in Z_PROVIDE_DATA.

When Z_AI decides to answer with Z_RSP, the following 8 rules applied:
1. Z_AI response contains only text, strictly compatible with Z_RSP:
2. Z_RSP format is completely defined by Z_RSP_FORMAT and Z_RSP_TEMPLATE.
3. Z_RSP_FORMAT is placed right between words Z_RSP_FORM_START and Z_RSP_FORM_END.
4. Z_RSP_FORMAT is totally defines syntax format of Z_RSP and used for automatic deserialization of Z_RSP.
5. Z_RSP_TEMPLATE is placed right between words Z_RSP_TEMP_START and Z_RSP_TEMP_END.
6. Z_RSP_TEMPLATE is totally defines logic format of Z_RSP and used for automatic deserialization of Z_RSP.
7. In answer Z_AI not uses any other symbols or words before or after Z_RSP, which may interfere with deserialization of Z_RSP_FORMAT and Z_RSP_TEMPLATE.
8. In answer Z_AI is as brief as possible, do not engages in any reasoning and only fills Z_RSP structure according to Z_RSP_FORMAT and Z_RSP_TEMPLATE,
places in the appropriate keys and arrays those values that corresponds to the provided data.

Z_RSP_FORM_START
%s
Z_RSP_FORM_END

Z_RSP_TEMP_START
%s
Z_RSP_TEMP_END

Z_DIALOG_START
`, agent.zRspFormatPrompt, zRspTemplateStr)

	agent.sysPrompt = "You are an AI interviewer. Ensure you are collected all the needed data from user to give complete answer."

	fmt.Printf("basicPrompt=%s\n", agent.basicPrompt)
	agent.zDialog = agent.basicPrompt
	return agent
}

// Run starts the interactive loop. It blocks until the context is cancelled or the process is terminated.
func (agent *AgentInterviewer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Print("\nПешы: ")
		userInput, _ := agent.reader.ReadString('\n')
		agent.zDialog = fmt.Sprintf("%s\n%s\n%s\n%s\n", agent.zDialog, agent.zProvideDataStart, strings.TrimSpace(userInput), agent.zProvideDataEnd)
		fmt.Printf("after provide zDialog=%s\n", agent.zDialog)

		escaped, err := json.Marshal(agent.zDialog)
		if err != nil {
			return err
		}

		resp, err := agent.client.CreateChatCompletion(
			context.Background(),
			openrouter.ChatCompletionRequest{
				Model: "deepseek/deepseek-chat-v3-0324:free",
				Messages: []openrouter.ChatCompletionMessage{
					{Role: openrouter.ChatMessageRoleSystem, Content: openrouter.Content{Text: agent.sysPrompt}},
					{Role: openrouter.ChatMessageRoleUser, Content: openrouter.Content{Text: string(escaped)}},
				},
			},
		)
		if err != nil {
			return err
		}

		respStr := resp.Choices[0].Message.Content.Text
		if strings.Contains(respStr, agent.zCollectDataStart) && strings.Contains(respStr, agent.zCollectDataEnd) {
			fmt.Printf("zCollectData respStr=%s\n", respStr)
			agent.zDialog = fmt.Sprintf("%s\n%s\n", agent.zDialog, respStr)
			fmt.Printf("after collect zDialog=%s\n", agent.zDialog)
			continue
		}

		if strings.Contains(respStr, agent.zRspStart) && strings.Contains(respStr, agent.zRspEnd) {
			respStrCut, err := cutN2(respStr, agent.zRspStart, agent.zRspEnd)
			if err != nil {
				return err
			}
			fmt.Printf("zRsp respStrCut=%s\n", respStrCut)

			var structuredRsp ZRsp
			if err := json.Unmarshal([]byte(respStrCut), &structuredRsp); err != nil {
				return err
			}
			fmt.Printf("structuredRsp=%v\n", structuredRsp)

			// Send to inspector and wait for inspection
			if agent.inspector != nil {
				if err := agent.inspector.Inspect(ctx, structuredRsp); err != nil {
					return err
				}
			}
			agent.zDialog = agent.basicPrompt
		}

		// Keep the same fallback/reset behavior as 1agent1user.go
		fmt.Printf("neither zRsp or zCollectData, reset; respStr=%s\n", respStr)
		agent.zDialog = agent.basicPrompt
	}
}
