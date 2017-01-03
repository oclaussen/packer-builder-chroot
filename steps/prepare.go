package steps

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/oclaussen/packer-builder-chroot/communicator"
	"strings"
)

type Prepare struct {
	ImageName string
	ImageSize string

	device    string
}

func (s *Prepare) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Preparing the image file...")

	image, err := filepath.Abs(s.ImageName)
	if err != nil {
		err := fmt.Errorf("Error preparing image file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	createCommand := fmt.Sprintf("dd if=/dev/zero of=%s bs=%s seek=1 count=0", image, s.ImageSize)
	if _, err := communicator.RunCommand(createCommand, state, nil); err != nil {
		return multistep.ActionHalt
	}

	ui.Say("Creating the loopback device...")

	setupCommand := fmt.Sprintf("losetup --find --show %s", image)
	device, err := communicator.RunCommand(setupCommand, state, nil)
	if err != nil {
		return multistep.ActionHalt
	}

	s.device = strings.TrimSpace(device)
	state.Put("device", s.device)
	state.Put("image", s.ImageName)
	return multistep.ActionContinue
}

func (s *Prepare) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	if s.device == "" {
		return
	}

	ui.Say("Detaching the loopback device...")
	detachCommand := fmt.Sprintf("losetup --detach %s", s.device)
	if _, err := communicator.RunCommand(detachCommand, state, nil); err != nil {
		return
	}

	s.device = ""
}
