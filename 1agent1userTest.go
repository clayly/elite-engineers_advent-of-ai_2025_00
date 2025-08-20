package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/revrost/go-openrouter"
)

func Run1Agent1UserTest() {
	zProvideDataStart := "Z_PROVIDE_DATA_START"
	zProvideDataEnd := "Z_PROVIDE_DATA_END"
	zCollectDataStart := "Z_COLLECT_DATA_START"
	zCollectDataEnd := "Z_COLLECT_DATA_END"
	zRspStart := "Z_RSP_START"
	zRspEnd := "Z_RSP_END"
	zRspFormat := "python code"
	zRspFormatPrompt := fmt.Sprintf("exact string %s, right after that valid %s, right after that exact string %s", zRspStart, zRspFormat, zRspEnd)

	fmt.Printf("zRspFormatPrompt=%s\n", zRspFormatPrompt)

	basicPrompt := fmt.Sprintf(
		`
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

When Z_AI decides to answer with Z_RSP, the following 3 rules applied:
1. Z_AI response contains only text, strictly compatible with Z_RSP:
2. Z_RSP format is completely defined by Z_RSP_FORMAT.
3. Z_RSP_FORMAT is placed right between words Z_RSP_FORM_START and Z_RSP_FORM_END.

Z_RSP_FORM_START
%s
Z_RSP_FORM_END

Z_DIALOG_START
`,
		zRspFormatPrompt)

	fmt.Printf("basicPrompt=%s\n", basicPrompt)

	llmToken := os.Getenv("OPENROUTER_API_KEY")
	if llmToken == "" {
		log.Fatal("export OPENROUTER_API_KEY first")
	}
	llmClient := openrouter.NewClient(llmToken)

	codeToTest, err := readFileToString("python-function.py")
	if err != nil {
		log.Fatalf("failed to read code: %v", err)
	}
	llmReqStr := basicPrompt + "\n" + strings.TrimSpace("write test for the python code below") + "\n" + codeToTest

	llmReqStrEscaped, err := json.Marshal(llmReqStr)
	if err != nil {
		log.Fatalf("failed to marshal text tp json string: %v", err)
	}

	resp, err := llmClient.CreateChatCompletion(
		context.Background(),
		openrouter.ChatCompletionRequest{
			//Model: "deepseek/deepseek-chat-v3-0324:free",
			Model: "qwen/qwen3-coder:free",
			Messages: []openrouter.ChatCompletionMessage{
				{Role: openrouter.ChatMessageRoleUser, Content: openrouter.Content{Text: string(llmReqStrEscaped)}},
			},
		},
	)

	if err != nil {
		log.Fatalf("llm err: %v", err)
	}

	respText := resp.Choices[0].Message.Content.Text
	fmt.Printf("llm rsp: %s\n", respText)

	writeStringToFile("python-test.py", respText)
}

func readFileToString(path string) (string, error) {
	// Открываем файл
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("не удалось открыть файл: %w", err)
	}
	// Обязательно закрываем файл после завершения функции
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("ошибка закрытия файла %v", err)
		}
	}(file)

	// Читаем весь файл в память
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении файла: %w", err)
	}

	// Преобразуем байты в строку
	return string(data), nil
}

func writeStringToFile(path string, content string) error {
	// os.WriteFile создаст файл, если его нет, или перезапишет существующий
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("не удалось записать файл: %w", err)
	}
	return nil
}
