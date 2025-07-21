package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	// fake import for commands to run their init() functions
	_ "github.com/lepinkainen/lambdabot/command"
	"github.com/lepinkainen/lambdabot/lambda"

	awslambda "github.com/aws/aws-lambda-go/lambda"
)

func main() {
	if os.Getenv("RUNMODE") == "stdout" {
		runLocal()
	} else {
		awslambda.Start(lambda.HandleRequest)
	}
}

func runLocal() {
	var cmd lambda.Command
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON input: %v\n", err)
		os.Exit(1)
	}

	result, err := lambda.HandleRequest(context.Background(), cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error handling request: %v\n", err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON output: %v\n", err)
		os.Exit(1)
	}
}
