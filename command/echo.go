package command

import "github.com/lepinkainen/lambdabot/lambda"

// Echo echoes the arguments back
func Echo(args string) (string, error) {

	return args, nil
}

func init() {
	lambda.RegisterHandler("echo", Echo)
}
