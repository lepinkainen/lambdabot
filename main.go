package main

import (
	// fake import for commands to run their init() functions
	_ "github.com/lepinkainen/lambdabot/command"
	"github.com/lepinkainen/lambdabot/lambda"

	awslambda "github.com/aws/aws-lambda-go/lambda"
)

func main() {
	awslambda.Start(lambda.HandleRequest)
}
