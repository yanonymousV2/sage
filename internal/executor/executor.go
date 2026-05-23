package executor

import (
	"os/exec"
)

type CommandResult struct {
	Command string
	Output  string
	Error   string
}

func Run(command string) CommandResult {
	if command == "" {
		return CommandResult{Command: command, Error: "empty command"}
	}

	cmd := exec.Command("sh", "-c", command)
	
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
