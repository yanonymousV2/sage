package executor

import (
	"os/exec"
	"strings"
)

type CommandResult struct {
	Command string
	Output  string
	Error   string
}

func Run(command string) CommandResult {

	parts := strings.Fields(command)

	if len(parts) == 0 {
		return CommandResult{
			Command: command,
			Output:  "",
			Error:   "empty command",
		}
	}

	cmd := exec.Command(parts[0], parts[1:]...)

	output, err := cmd.CombinedOutput()

	result := CommandResult{
		Command: command,
		Output:  string(output),
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}
