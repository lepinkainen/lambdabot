# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Lambdabot is a Go-based AWS Lambda function that implements Pyfibot command functionality. It provides a command-based system where external clients can invoke specific commands through Lambda function calls.

## Critical Development Workflow

**ALWAYS follow this sequence when making changes:**

1. Run `goimports -w .` on modified Go files (not `gofmt` - `goimports` includes import management)
2. Run `task build` to ensure code builds successfully (includes tests, linting, and cross-compilation)
3. Write basic unit tests for new functionality - tests are in `*_test.go` files alongside implementation

## Build System & Commands

This project uses Taskfile for build management. Key commands:

- `task` or `task build` - **Primary build command** (runs tests → cross-compiles for Linux ARM64 → creates zip)
- `task build-local` - Build for local testing (depends on tests passing)
- `task test` - Run tests with coverage and vet
- `task test-ci` - CI tests (excludes files with `ci` build tag)
- `task lint` - Run golangci-lint
- `task clean` - Remove build artifacts
- `task publish` - Deploy to AWS Lambda (requires AWS credentials, runs lint + build)
- `task upgrade-deps` - Upgrade all dependencies

**Build Dependencies:** All build tasks depend on tests passing first. The build creates a Linux ARM64 binary and zip package.

## Architecture

### Core Components & Data Flow

1. **main.go** - Entry point with blank import to trigger command `init()` functions
2. **lambda/main.go** - Central handler registry and request processor
3. **command/** - Individual command implementations (auto-registered via `init()`)

### Command Registration System

**Critical Pattern:** Commands auto-register through Go's `init()` system:

- **main.go** uses blank import `_ "github.com/lepinkainen/lambdabot/command"` to trigger all command `init()` functions
- **lambda/main.go** maintains `handlerFunctions` map and processes requests via `HandleRequest()`
- Commands are matched by **exact string comparison** - first match wins, others ignored

### Request Processing Flow

```
AWS Lambda → HandleRequest() → handlerFunctions[cmd.Command] → Command Function → Response
```

The `Command` struct (lambda/main.go:14-21) defines the request/response format with fields: User, Source, Command, Arguments, Result.

### Adding New Commands

**Follow this exact pattern** (see command/echo.go for reference):

1. Create `command/yourcommand.go`
2. Implement function: `func YourCommand(args string) (string, error)`
3. Add `init()` function: `lambda.RegisterHandler("commandname", YourCommand)`
4. No import needed in main.go - blank import handles it

Example (minimal working command):
```go
package command

import "github.com/lepinkainen/lambdabot/lambda"

func MyCommand(args string) (string, error) {
    return "processed: " + args, nil
}

func init() {
    lambda.RegisterHandler("mycommand", MyCommand)
}
```

## Development Guidelines

- **Language**: Go only - no Python or other languages
- **Code Formatting**: Use `goimports -w .` (not `gofmt`) - includes automatic import management
- **Testing**: Write basic unit tests in `*_test.go` files - focus on critical functionality, not 100% coverage
- **API Keys**: Use environment variables for external API keys (see existing commands for patterns)
- **Logging**: Structured JSON logging via logrus (configured in lambda/main.go:52-62)
- **Error Handling**: Return meaningful errors from command functions - errors are logged automatically

## Code Patterns & Conventions

**External API Integration** (see command/tvmaze.go, command/openweathermap.go):
- Use standard `net/http` for HTTP requests
- Define response structs with proper JSON tags
- Use `any` type for dynamic/variable fields (runtime, summary)
- Handle API errors gracefully with meaningful messages

**Environment Variables**: Commands needing API keys use `os.Getenv()` pattern
**Testing**: Each command has corresponding `*_test.go` with success/error test cases

## Dependencies

Current key dependencies:
- `github.com/aws/aws-lambda-go` - AWS Lambda runtime
- `github.com/sirupsen/logrus` - Structured JSON logging (configured for Lambda)
- `github.com/pkg/errors` - Enhanced error handling
- `github.com/dustin/go-humanize` - Human-friendly formatting

**Adding Dependencies**: Justify new third-party dependencies - prefer standard library when possible.

## Testing Strategy

### Unit Testing
```bash
task test        # Local testing with all packages
task test-ci     # CI testing (excludes files with 'ci' build tag)
```

**Test Pattern**: Each command has a `*_test.go` file testing both success and error scenarios.
**CI Integration**: Use `//go:build !ci` tag to exclude tests requiring external APIs from CI runs.

### Local Integration Testing

**RUNMODE Parameter**: Test complete Lambda handler locally without AWS deployment.

```bash
# Build local binary
go build -o lambdabot

# Test echo command
echo '{"command":"echo","args":"test message"}' | RUNMODE=stdout ./lambdabot

# Test with full Command struct
echo '{"user":"dev","source":"local","command":"echo","args":"Hello"}' | RUNMODE=stdout ./lambdabot
```

**RUNMODE=stdout behavior**:
- Reads JSON from stdin (Command struct format)
- Calls `HandleRequest()` with full Lambda context
- Outputs JSON response to stdout
- Logs go to stderr (can be redirected with `2>/dev/null`)
- Tests complete request/response cycle including JSON marshaling

**Local Testing Benefits**:
- No AWS deployment needed for integration testing
- Test complete Lambda handler flow
- Validate JSON request/response format
- Debug command registration and execution
- Fast iteration during development

**Integration Test Script**: Run `./test_local.sh` for automated local testing of RUNMODE functionality.

## Deployment Architecture

**Lambda Package Creation**:
- Cross-compiles to Linux ARM64 (`GOOS=linux GOARCH=arm64`)
- Binary named `bootstrap` (Lambda runtime requirement)
- Packaged as zip file for deployment
- Git commit hash embedded via build flags