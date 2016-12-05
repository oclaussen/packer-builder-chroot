package steps

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/oclaussen/packer-builder-chroot/communicator"
)

type postMountData struct {
	Device    string
	MountPath string
}

type PostMount struct {
	Commands []string
}

func (s *PostMount) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if len(s.Commands) == 0 {
		return multistep.ActionContinue
	}

	data := &postMountData{
		Device:    state.Get("device").(string),
		MountPath: state.Get("mount_path").(string),
	}

	ui.Say("Running post-mount commands...")
	for _, command := range s.Commands {
		if _, err := communicator.RunCommand(command, state, data); err != nil {
			return multistep.ActionHalt
		}
	}
	return multistep.ActionContinue
}

func (s *PostMount) Cleanup(state multistep.StateBag) {}
