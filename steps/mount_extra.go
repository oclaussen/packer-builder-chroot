package steps

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
	"github.com/oclaussen/packer-builder-chroot/communicator"
)

type MountExtra struct {
	ChrootMounts [][]string

	mounts       []string
}

func (s *MountExtra) Run(state multistep.StateBag) multistep.StepAction {
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packer.Ui)

	s.mounts = make([]string, 0, len(s.ChrootMounts))

	ui.Say("Mounting additional paths within the chroot...")
	for _, mountInfo := range s.ChrootMounts {
		innerPath := mountPath + mountInfo[2]

		if err := os.MkdirAll(innerPath, 0755); err != nil {
			err := fmt.Errorf("Error creating mount directory: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		flags := "-t " + mountInfo[0]
		if mountInfo[0] == "bind" {
			flags = "--bind"
		}

		ui.Message(fmt.Sprintf("Mounting: %s", mountInfo[2]))
		mountCommand := fmt.Sprintf("mount %s %s %s", flags, mountInfo[1], innerPath)
		if _, err := communicator.RunCommand(mountCommand, state, nil); err != nil {
			return multistep.ActionHalt
		}

		s.mounts = append(s.mounts, innerPath)
	}

	return multistep.ActionContinue
}

func (s *MountExtra) Cleanup(state multistep.StateBag) {
	if s.mounts == nil {
		return
	}

	for len(s.mounts) > 0 {
		var path string
		lastIndex := len(s.mounts) - 1
		path, s.mounts = s.mounts[lastIndex], s.mounts[:lastIndex]

		grepCommand := fmt.Sprintf("grep %s /proc/mounts", path)
		if _, err := communicator.RunCommand(grepCommand, state, nil); err != nil {
			continue;
		}

		umountCommand := fmt.Sprintf("umount %s", path)
		if _, err := communicator.RunCommand(umountCommand, state, nil); err != nil {
			return
		}
	}

	s.mounts = nil
	return
}
