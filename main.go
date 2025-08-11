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
After the word Z_REQ I will send you a request.

You response should contain only text, strictly compatible with Z_RSP.

Z_RSP is completely defined by Z_RSP_FORMAT and Z_RSP_TEMPLATE.

Z_RSP_FORMAT is specified between the words Z_RSP_FORM_START and Z_RSP_FORM_END.
Z_RSPF_FORMAT is totally defines syntax format of Z_RSP and used for automatic deserialization of Z_RSP.

Z_RSP_TEMPLATE is specified between the words Z_RSP_TEMP_START and Z_RSP_TEMP_END.
Z_RSP_TEMPLATE is totally defines logic format of Z_RSP and used for automatic deserialization of Z_RSP.

Do not use any other symbols or words before or after Z_RSP, which may interfere with deserialization of Z_RSP_FORMAT and Z_RSP_TEMPLATE.

Be as brief as possible, do not engage in any reasoning in response, only fill the Z_RSP structure according to Z_RSP_FORMAT and Z_RSP_TEMPLATE,
placing in the appropriate keys or arrays those values that correspond to the response.

Z_RSP_FORM_START
%s
Z_RSP_FORM_END

Z_RSP_TEMP_START
%s
Z_RSP_TEMP_END

Z_REQ
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
