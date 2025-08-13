# elite-engineers_advent-of-ai_2025_00

Small Go CLI that drives a goal-oriented dialog with an LLM via OpenRouter. It alternates between asking for clarifying data and, when ready, emitting a final structured JSON response that matches a predefined schema.

## Prerequisites
- Go 1.24.x (toolchain go1.24.6)
- An OpenRouter API key exported as an environment variable:
  - macOS/Linux: `export OPENROUTER_API_KEY=your_openrouter_api_key`
  - Windows (PowerShell): `$Env:OPENROUTER_API_KEY="your_openrouter_api_key"`

## Models & Client
The project uses github.com/revrost/go-openrouter and an example free model `deepseek/deepseek-chat-v3-0324:free`.

## Build and Run (2 agents, 1 user)

The main entry point invokes the 2-agents-1-user flow (Run2Agents1User).

1) Build the binary
- macOS/Linux/Windows (with Go in PATH):
  - go build -o advent

2) Set your OpenRouter API key
- macOS/Linux (bash/zsh):
  - export OPENROUTER_API_KEY=your_openrouter_api_key
- Windows (PowerShell):
  - $Env:OPENROUTER_API_KEY="your_openrouter_api_key"

3) Run the app
- macOS/Linux:
  - ./advent
- Windows:
  - .\advent.exe

What to expect
- The program will prompt with "Пешы:". Type your initial input to set the dialog direction.
- The app alternates between collecting data (responses will include Z_COLLECT_DATA markers) and eventually returns a structured JSON block between Z_RSP_START and Z_RSP_END that matches the internal ZRsp schema.
- The structured JSON will also be sent to the inspector agent for a brief acknowledgment.
