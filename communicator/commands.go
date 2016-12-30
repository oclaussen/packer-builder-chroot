package communicator

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/multistep"
)

type CommandWrapper func(string) (string, error)

func RunCommand(command string, state multistep.StateBag, data interface{}) (string, error) {
	ui := state.Get("ui").(packer.Ui)
	cmdWrapper := state.Get("wrappedCommand").(CommandWrapper)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	comm := &Shell{
		CmdWrapper: cmdWrapper,
		Data:       data,
	}

	cmd := &packer.RemoteCmd{
		Command: command,
		Stdout:  stdout,
		Stderr:  stderr,
	}


	if err := cmd.StartWithUi(comm, ui); err != nil {
		err = fmt.Errorf("Error executing command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return stderr.String(), err
	}

	if cmd.ExitStatus != 0 {
		err := fmt.Errorf("Received non-zero exit code %d from command: %s", cmd.ExitStatus, command)
		state.Put("error", err)
		ui.Error(err.Error())
		return stderr.String(), err
	}

	return stdout.String(), nil
}