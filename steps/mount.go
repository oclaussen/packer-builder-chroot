package steps

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"github.com/oclaussen/packer-builder-chroot/communicator"
)

type mountPathData struct {
	Device string
}

type Mount struct {
	MountPath       string
	MountOptions    []string
	MountPartitions [][]string

	finalMounts	[]string
}

func (s *Mount) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	device := state.Get("device").(string)

	s.finalMounts = make([]string, 0, len(s.MountPartitions))

	data := &mountPathData{Device: filepath.Base(device)}

	mountPath, err := interpolate.Render(s.MountPath, &interpolate.Context{Data: data})
	if err != nil {
		err := fmt.Errorf("Error preparing mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	log.Printf("Mount path: %s", mountPath)

	for _, mountInfo := range s.MountPartitions {
		innerPath, err := filepath.Abs(mountPath + mountInfo[1])
		if err != nil {
			err := fmt.Errorf("Error preparing mount directory: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if err := os.MkdirAll(innerPath, 0755); err != nil {
			err := fmt.Errorf("Error creating mount directory: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Say("Mounting the root device...")

		opts := ""
		if len(s.MountOptions) > 0 {
			opts = "-o " + strings.Join(s.MountOptions, " -o ")
		}

		mountCommand := fmt.Sprintf("sudo mount %s %sp%s %s", opts, device, mountInfo[0], innerPath)
		if _, err := communicator.RunCommand(mountCommand, state, data); err != nil {
			return multistep.ActionHalt
		}

		s.finalMounts = append(s.finalMounts, innerPath)
	}

	state.Put("mount_path", mountPath)
	return multistep.ActionContinue
}

func (s *Mount) Cleanup(state multistep.StateBag) {
	if s.finalMounts == nil {
		return
	}

	for len(s.finalMounts) > 0 {
		var path string
		lastIndex := len(s.finalMounts) - 1
		path, s.finalMounts = s.finalMounts[lastIndex], s.finalMounts[:lastIndex]

		grepCommand := fmt.Sprintf("grep %s /proc/mounts", path)
		if _, err := communicator.RunCommand(grepCommand, state, nil); err != nil {
			continue;
		}

		umountCommand := fmt.Sprintf("sudo umount %s", path)
		if _, err := communicator.RunCommand(umountCommand, state, nil); err != nil {
			return
		}
	}

	s.finalMounts = nil
	return
}
