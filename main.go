package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/revrost/go-openrouter"
)

type ZRspItem struct {
	ItemType    string `json:"itemType"`
	ItemName    string `json:"itemName"`
	Value1Name  string `json:"value1Name"`
	Value1Units string `json:"value1Units"`
	Value1      string `json:"value1"`
}

type ZRsp struct {
	Items []ZRspItem `json:"items"`
}

func main() {
	zRspStart := "zRspStart"
	zRspEnd := "zRspEnd"
	zRspFormat := "JSON"
	zRspFormatPrompt := fmt.Sprintf("exact string %s, right after that valid %s, right after that exact string %s", zRspStart, zRspFormat, zRspEnd)

	fmt.Printf("zRspFormatPrompt=%s\n", zRspFormatPrompt)

	zRspTemplate := ZRsp{
		Items: []ZRspItem{
			{
				ItemType:    "car",
				ItemName:    "TT-34",
				Value1Name:  "cost",
				Value1Units: "byn",
				Value1:      "1000000",
			},
			{
				ItemType:    "bullet",
				ItemName:    "7.62x39mm",
				Value1Name:  "speed",
				Value1Units: "km/h",
				Value1:      "360",
			},
		},
	}

	zRspTemplateStr, err := json.Marshal(zRspTemplate)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("zRspTemplateStr=%s\n", string(zRspTemplateStr))

	basicPrompt := fmt.Sprintf(
		`
This is dialog (designated as Z_DIALOG) between me (the user, designated as Z_USER) and you (the AI, designated as Z_AI).

Z_AI can only do two things in Z_DIALOG:
1. Z_AI can give an answer (designated as Z_RSP), which will finish Z_DIALOG.
1. Z_AI can collect data from Z_USER with clarifying questions (designated as Z_COLLECT_DATA) in order to qualitatively fill out Z_RSP.

Z_AI main and only goal in this dialog is to fill Z_RSP, format of which is described below.

Z_USER do two things in Z_DIALOG:
1. Z_USER can provide additional data with clarifying answers (designated as Z_PROVIDE_DATA). 
1. Z_USER determines direction of Z_DIALOG and therefore content of Z_RSP with his first Z_PROVIDE_DATA.

All Z_COLLECT_DATA placed between strings Z_COLLECT_DATA_START_{N} and Z_COLLECT_DATA_END_{N}, where {N} is number of this Z_COLLECT_DATA in Z_DIALOG. 
All Z_PROVIDE_DATA placed between strings Z_PROVIDE_DATA_START_{N} and Z_PROVIDE_DATA_END_{N}, where {N} is number of this Z_PROVIDE_DATA in Z_DIALOG. 

Z_AI finishes Z_DIALOG with Z_RSP in two occasions:
1. When Z_AI decides that enough data is provided.
2. When Z_AI decides that Z_USER is not providing relative data anymore in Z_PROVIDE_DATA.

When Z_AI decides to give answer, Z_AI response contains only text, strictly compatible with Z_RSP:
1. Z_RSP format is completely defined by Z_RSP_FORMAT and Z_RSP_TEMPLATE.
2. Z_RSP_FORMAT is placed right between strings Z_RSP_FORM_START and Z_RSP_FORM_END.
3. Z_RSP_FORMAT is totally defines syntax format of Z_RSP and used for automatic deserialization of Z_RSP.
4. Z_RSP_TEMPLATE is placed right between strings Z_RSP_TEMP_START and Z_RSP_TEMP_END.
5. Z_RSP_TEMPLATE is totally defines logic format of Z_RSP and used for automatic deserialization of Z_RSP.
6. In answer Z_AI not uses any other symbols or words before or after Z_RSP, which may interfere with deserialization of Z_RSP_FORMAT and Z_RSP_TEMPLATE.
7. In answer Z_AI is as brief as possible, do not engages in any reasoning and only fills Z_RSP structure according to Z_RSP_FORMAT and Z_RSP_TEMPLATE,
places in the appropriate keys and arrays those values that corresponds to the provided data.

Z_RSP_FORM_START
%s
Z_RSP_FORM_END

Z_RSP_TEMP_START
%s
Z_RSP_TEMP_END

`,
		zRspFormatPrompt,
		zRspTemplateStr)

	fmt.Printf("basicPrompt=%s\n", basicPrompt)

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("export OPENROUTER_API_KEY first")
	}

	client := openrouter.NewClient(apiKey)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nПешы: ")
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)

		escaped, err := json.Marshal(basicPrompt + " " + userInput)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openrouter.ChatCompletionRequest{
				Model: "deepseek/deepseek-chat-v3-0324:free",
				Messages: []openrouter.ChatCompletionMessage{
					{
						Role:    openrouter.ChatMessageRoleUser,
						Content: openrouter.Content{Text: string(escaped)},
					},
				},
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}

		respStr := resp.Choices[0].Message.Content.Text
		fmt.Printf("respStr=%s\n", respStr)

		respStrCut, err := cutN2(respStr, zRspStart, zRspEnd)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("respStrCut=%s\n", respStrCut)

		var structuredRsp ZRsp
		if err := json.Unmarshal([]byte(respStrCut), &structuredRsp); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("structuredRsp=%v\n", structuredRsp)
	}
}

func cutN2(src, n1, n3 string) (string, error) {
	if n1 == "" || n3 == "" {
		return "", errors.New("n1 и n3 не должны быть пустыми")
	}

	// позиция первого n1
	i1 := strings.Index(src, n1)
	if i1 == -1 {
		return "", errors.New("n1 не найден")
	}
	start := i1 + len(n1)

	// позиция последнего n3
	i3 := strings.LastIndex(src, n3)
	if i3 == -1 || i3 < start {
		return "", errors.New("n3 не найден или расположен до n1")
	}

	return src[start:i3], nil
}
