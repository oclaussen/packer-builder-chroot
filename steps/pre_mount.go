package steps

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/oclaussen/packer-builder-chroot/communicator"
)

type preMountData struct {
	Device string
}

type PreMount struct {
	Commands []string
}

func (s *PreMount) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if len(s.Commands) == 0 {
		return multistep.ActionContinue
	}

	data := &preMountData{
		Device: state.Get("device").(string),
	}

	ui.Say("Running device setup commands...")
	for _, command := range s.Commands {
		if _, err := communicator.RunCommand(command, state, data); err != nil {
			return multistep.ActionHalt
		}
	}
	return multistep.ActionContinue
}

func (s *PreMount) Cleanup(state multistep.StateBag) {}
