# Project Guidelines

## Project Overview
This repository contains a small Go command-line application that interacts with the OpenRouter API to conduct a goal-driven dialog with a large language model (LLM). The application alternates between:
- Collecting clarifying data from the user (content placed between Z_COLLECT_DATA_START and Z_COLLECT_DATA_END), and
- Producing a final structured JSON response (Z_RSP) delimited by Z_RSP_START and Z_RSP_END that matches a predefined JSON schema and template and can be deserialized.

### Key Features
- Structured prompt embedding a response format descriptor and a JSON template (ZRsp) used for parsing.
- Interactive loop that appends user input to the dialog context and requests an LLM response via OpenRouter.
- Automatic detection of Z_COLLECT_DATA vs. Z_RSP and JSON unmarshalling of the final response.

### Project Structure
- main.go — CLI entry point and dialog orchestration.
- go.mod / go.sum — module metadata and dependencies.
- README.md — repository title placeholder.
- .junie/guidelines.md — this document.

### Requirements
- Go 1.24.x (toolchain go1.24.6 as specified in go.mod).
- Environment variable OPENROUTER_API_KEY must be set to a valid OpenRouter API key.

### Build and Run
1. Build the binary:
   - go build -o advent
2. Export your API key:
   - export OPENROUTER_API_KEY=your_openrouter_api_key
3. Run the app:
   - ./advent
4. The program will prompt for input each cycle and print debug output including the evolving dialog and any structured responses.

### Dependencies
- github.com/revrost/go-openrouter v0.2.0 (OpenRouter API client)
- Indirect: zerolog, x/sys, and others as listed in go.mod.

### Testing and Verification
- No automated tests are present.
- Manual check: run the binary, provide a few sample inputs, and confirm that:
  - When the model requests more information, the response contains Z_COLLECT_DATA markers.
  - When sufficient data is provided, the model returns a JSON block between Z_RSP_START and Z_RSP_END that unmarshals into the ZRsp struct without errors.

### Coding Style and Notes
- Use go fmt and idiomatic Go error handling.
- Keep changes minimal; do not alter sentinel markers or dialog flow unless necessary.
- Do not commit secrets; use OPENROUTER_API_KEY via environment variables.
