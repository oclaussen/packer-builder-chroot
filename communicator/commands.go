package communicator

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/template/interpolate"
)

type CommandWrapper func(string) (string, error)

func RunCommand(command string, state multistep.StateBag, data interface{}) (string, error) {
	ui := state.Get("ui").(packer.Ui)
	cmdWrapper := state.Get("wrappedCommand").(CommandWrapper)

	command, err := interpolate.Render(command, &interpolate.Context{Data: data})
	if err != nil {
		err = fmt.Errorf("Error interpolating: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return "", err
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	comm := &Shell{CmdWrapper: cmdWrapper}

	cmd := &packer.RemoteCmd{
		Command: command,
		Stdout:  stdout,
		Stderr:  stderr,
	}

	ui.Message(fmt.Sprintf("Executing command: %s", command))

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