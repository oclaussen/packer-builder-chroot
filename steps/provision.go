package steps

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"github.com/oclaussen/packer-builder-chroot/communicator"
)

type Provision struct {
	mounts []string
}

func (s *Provision) Run(state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	ui := state.Get("ui").(packer.Ui)

	comm := &communicator.Chroot{
		Chroot:     state.Get("mount_path").(string),
		CmdWrapper: state.Get("wrappedCommand").(communicator.CommandWrapper),
	}

	log.Println("Running the provision hook")
	if err := hook.Run(packer.HookProvision, ui, comm, nil); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *Provision) Cleanup(state multistep.StateBag) {}
