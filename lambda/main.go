package lambda

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	handlerFunctions = make(map[string]func(string) (string, error))
)

// Command is the query and response to commands
type Command struct {
	User      string `json:"user"`
	Source    string `json:"source"`
	Command   string `json:"command"`
	Arguments string `json:"args"`
	Result    string `json:"result"`
}

type handlerFunc func(string) (string, error)

// RegisterHandler adds the given url parser and command to the map of handlers
func RegisterHandler(command string, function handlerFunc) {
	handlerFunctions[command] = function
}

// HandleRequest is the function entry point
func HandleRequest(ctx context.Context, cmd Command) (Command, error) {

	log.Infof("Handling %v", cmd)

	// NOTE: only the first matching command will be run
	for pattern, handler := range handlerFunctions {
		// No match, skip
		if pattern != cmd.Command {
			continue
		}
		log.Infof("Running command %v", cmd)
		res, err := handler(cmd.Arguments)
		cmd.Result = res
		return cmd, err
	}

	// No handler matched
	cmd.Result = fmt.Sprintf("Unknown command: %s", cmd.Command)
	return cmd, nil
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}
