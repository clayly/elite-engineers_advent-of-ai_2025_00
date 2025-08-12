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
	zProvideDataStart := "Z_PROVIDE_DATA_START"
	zProvideDataEnd := "Z_PROVIDE_DATA_END"
	zCollectDataStart := "Z_COLLECT_DATA_START"
	zCollectDataEnd := "Z_COLLECT_DATA_END"
	zRspStart := "Z_RSP_START"
	zRspEnd := "Z_RSP_END"
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
			{
				ItemType:    "action",
				ItemName:    "deleting folder in Linux",
				Value1Name:  "bash command",
				Value1Units: "bash code",
				Value1:      "sudo rm -rf {folder_name}",
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
There is dialog (named Z_DIALOG) between me (the user, named Z_USER) and you (the AI, named Z_AI).
Z_DIALOG starts after word Z_DIALOG_START.

Z_AI can only do 2 things in Z_DIALOG:
1. Z_AI can give an answer (named Z_RSP, format of which is described below), which will finish Z_DIALOG.
2. Z_AI can collect data from Z_USER with clarifying questions (named Z_COLLECT_DATA) in order to qualitatively fill out Z_RSP with specific information.

All Z_COLLECT_DATA in Z_DIALOG placed between words Z_COLLECT_DATA_START and Z_COLLECT_DATA_END.

Z_AI main and only goal in Z_DIALOG is to collect enough specific data to fill Z_RSP.
Z_AI most ensure that it is collected enough specific data to fill Z_RSP.

Z_USER can only do 2 things in Z_DIALOG:
1. Z_USER can provide additional data with clarifying answers (named Z_PROVIDE_DATA). 
2. Z_USER determines direction of Z_DIALOG and therefore content of Z_RSP with his first Z_PROVIDE_DATA.

All Z_PROVIDE_DATA in Z_DIALOG placed between words Z_PROVIDE_DATA_START and Z_PROVIDE_DATA_END.

Z_AI finishes Z_DIALOG with Z_RSP in 2 occasions:
1. When Z_USER when the user clearly writes that he can't clarify anything more.

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

	zDialog := basicPrompt

	for {
		fmt.Print("\nПешы: ")
		userInput, _ := reader.ReadString('\n')
		zDialog = fmt.Sprintf(
			"%s\n%s\n%s\n%s\n",
			zDialog,
			zProvideDataStart,
			strings.TrimSpace(userInput),
			zProvideDataEnd)

		fmt.Printf("after provide zDialog=%s\n", zDialog)

		escaped, err := json.Marshal(zDialog)
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
			log.Fatal(err)
			return
		}

		respStr := resp.Choices[0].Message.Content.Text

		if strings.Contains(respStr, zCollectDataStart) && strings.Contains(respStr, zCollectDataEnd) {
			fmt.Printf("zCollectData respStr=%s\n", respStr)
			zDialog = fmt.Sprintf(
				"%s\n%s\n",
				zDialog,
				respStr)
			fmt.Printf("after collect zDialog=%s\n", zDialog)
			continue
		}

		if strings.Contains(respStr, zRspStart) && strings.Contains(respStr, zRspEnd) {
			respStrCut, err := cutN2(respStr, zRspStart, zRspEnd)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("zRsp respStrCut=%s\n", respStrCut)

			var structuredRsp ZRsp
			if err := json.Unmarshal([]byte(respStrCut), &structuredRsp); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("structuredRsp=%v\n", structuredRsp)
			zDialog = basicPrompt
		}

		fmt.Printf("neither zRsp or zCollectData, reset; respStr=%s\n", respStr)
		zDialog = basicPrompt
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
