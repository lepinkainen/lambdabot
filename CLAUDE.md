# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Lambdabot is a Go-based AWS Lambda function that implements Pyfibot command functionality. It provides a command-based system where external clients can invoke specific commands through Lambda function calls.

## Build System & Commands

This project uses Taskfile for build management. Common commands:

- `task` or `task build` - Build the Lambda function (cross-compiles for Linux ARM64)
- `task build-local` - Build for local testing
- `task test` - Run tests with coverage and vet
- `task lint` - Run golangci-lint
- `task clean` - Remove build artifacts
- `task publish` - Deploy to AWS Lambda (requires AWS credentials)
- `task upgrade-deps` - Upgrade all dependencies

## Architecture

### Core Components

1. **main.go** - Entry point that starts the AWS Lambda handler
2. **lambda/main.go** - Core Lambda handler with command registration system
3. **command/** - Individual command implementations

### Command System

The application uses a registration-based command system:

- Commands are registered in their `init()` functions using `lambda.RegisterHandler()`
- Each command implements the signature: `func(string) (string, error)`
- Commands are matched by exact string comparison in `HandleRequest()`
- Only the first matching command is executed

### Adding New Commands

To add a new command:

1. Create a new file in `command/` directory
2. Implement a function with signature `func(string) (string, error)`
3. Register it in an `init()` function: `lambda.RegisterHandler("commandname", YourFunction)`
4. Import the command package in main.go (usually handled by the blank import)

Example command structure:
```go
package command

import "github.com/lepinkainen/lambdabot/lambda"

func MyCommand(args string) (string, error) {
    // Implementation
    return "result", nil
}

func init() {
    lambda.RegisterHandler("mycommand", MyCommand)
}
```

## Development Guidelines

- **Language**: Go only
- **Testing**: Write unit tests for new functionality in `*_test.go` files
- **API Keys**: Use environment variables for external API keys
- **Logging**: Use structured JSON logging via logrus
- **Error Handling**: Return meaningful errors from command functions

## Dependencies

Key dependencies:
- `github.com/aws/aws-lambda-go` - AWS Lambda runtime
- `github.com/sirupsen/logrus` - Structured logging
- `github.com/redis/go-redis/v9` - Redis client (if used)

## Testing

Run tests after making changes:
```bash
task test
```

Individual command tests follow the pattern `command_test.go` and test both success and error cases.

## Deployment

The project builds a Lambda deployment package:
- Cross-compiles for Linux ARM64
- Creates a zip file with the `bootstrap` binary
- Deployment via `task publish` requires AWS credentials and proper Lambda function configuration