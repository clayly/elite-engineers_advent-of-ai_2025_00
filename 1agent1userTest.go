package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/revrost/go-openrouter"
)

func Run1Agent1UserTest() {
	zRspStart := "Z_RSP_START"
	zRspEnd := "Z_RSP_END"
	zRspFormat := "python code"
	zRspFormatPrompt := fmt.Sprintf("exact string %s, right after that valid %s, right after that exact string %s", zRspStart, zRspFormat, zRspEnd)

	fmt.Printf("zRspFormatPrompt=%s\n", zRspFormatPrompt)

	llmSystemPrompt := fmt.Sprintf(
		`
You are coding assistant (named Z_AI).

Z_AI can only do 1 thing: give an answer (named Z_RSP, format of which is described below), which will finish Z_DIALOG.

Z_AI response contains only text, strictly compatible with Z_RSP.

Z_RSP format is completely defined by Z_RSP_FORMAT.

Z_RSP_FORMAT is placed right between words Z_RSP_FORM_START and Z_RSP_FORM_END.

Z_RSP_FORM_START
%s
Z_RSP_FORM_END
`,
		zRspFormatPrompt)

	llmSystemPromptEscaped, err := json.Marshal(llmSystemPrompt)
	if err != nil {
		log.Fatalf("failed to marshal text tp json string: %v", err)
	}

	fmt.Printf("basicPrompt=%s\n", llmSystemPrompt)

	codeToTest, err := readFileToString("function_python.py")
	if err != nil {
		log.Fatalf("failed to read code: %v", err)
	}

	llmUserPromptStr := strings.TrimSpace("write test for the python code below") + "\n" + codeToTest

	llmUserPromptEscaped, err := json.Marshal(llmUserPromptStr)
	if err != nil {
		log.Fatalf("failed to marshal text tp json string: %v", err)
	}

	llmToken := os.Getenv("OPENROUTER_API_KEY")
	if llmToken == "" {
		log.Fatal("export OPENROUTER_API_KEY first")
	}
	llmClient := openrouter.NewClient(llmToken)

	resp, err := llmClient.CreateChatCompletion(
		context.Background(),
		openrouter.ChatCompletionRequest{
			Model: "moonshotai/kimi-k2:free",
			//Model: "deepseek/deepseek-chat-v3-0324:free",
			//Model: "qwen/qwen3-coder:free",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    openrouter.ChatMessageRoleSystem,
					Content: openrouter.Content{Text: string(llmSystemPromptEscaped)},
				},
				{
					Role:    openrouter.ChatMessageRoleUser,
					Content: openrouter.Content{Text: string(llmUserPromptEscaped)},
				},
			},
		},
	)

	if err != nil {
		log.Fatalf("llm err: %v", err)
	}

	respStr := resp.Choices[0].Message.Content.Text
	fmt.Printf("llm rsp: %s\n", respStr)

	codeOfTest, err := cutN2(respStr, zRspStart, zRspEnd)
	if err != nil {
		log.Fatalf("cur codeOfTest err: %v", err)
	}

	err = writeStringToFile("tmp/test_python.py", codeOfTest)
	if err != nil {
		log.Printf("ошибка записи файла с тестом: %v", err)
	}

	cmd := exec.Command(
		"docker",
		"build",
		"--progress=plain",
		"--no-cache",
		"-t",
		"advent-pytest",
		".",
		"--file",
		"Dockerfile-pytest",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start docker container: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("Failed to wait for docker container: %v", err)
	}
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
