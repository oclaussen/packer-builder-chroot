package steps

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"path/filepath"
	"github.com/oclaussen/packer-builder-chroot/communicator"
)

type CopyFiles struct {
	Files []string
}

func (s *CopyFiles) Run(state multistep.StateBag) multistep.StepAction {
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packer.Ui)

	if len(s.Files) == 0 {
		return multistep.ActionContinue
	}

	ui.Say("Copying files from host to chroot...")
	for _, path := range s.Files {
		ui.Message(path)
		chrootPath := filepath.Join(mountPath, path)
		log.Printf("Copying '%s' to '%s'", path, chrootPath)

		cpCmd := fmt.Sprintf("cp --remove-destination %s %s", path, chrootPath)
		if _, err := communicator.RunCommand(cpCmd, state, nil); err != nil {
			return multistep.ActionHalt
		}

		s.Files = append(s.Files, chrootPath)
	}

	return multistep.ActionContinue
}

func (s *CopyFiles) Cleanup(state multistep.StateBag) {
	if s.Files == nil {
		return
	}

	for _, file := range s.Files {
		log.Printf("Removing: %s", file)
		rmCmd := fmt.Sprintf("rm -f %s", file)
		if _, err := communicator.RunCommand(rmCmd, state, nil); err != nil {
			return
		}
	}

	s.Files = nil
}
