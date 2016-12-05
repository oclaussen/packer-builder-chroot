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
	MountOptions   []string
	MountPartition int
	MountPath      string

	finalMountPath string
}

func (s *Mount) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	device := state.Get("device").(string)

	data := &mountPathData{Device: filepath.Base(device)}

	mountPath, err := interpolate.Render(s.MountPath, &interpolate.Context{Data: data})
	if err != nil {
		err := fmt.Errorf("Error preparing mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	mountPath, err = filepath.Abs(mountPath)
	if err != nil {
		err := fmt.Errorf("Error preparing mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Mount path: %s", mountPath)

	if err := os.MkdirAll(mountPath, 0755); err != nil {
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

	mountCommand := fmt.Sprintf("sudo mount %s %s %s", opts, device, mountPath)
	if _, err := communicator.RunCommand(mountCommand, state, data); err != nil {
		return multistep.ActionHalt
	}

	s.finalMountPath = mountPath
	state.Put("mount_path", s.finalMountPath)
	return multistep.ActionContinue
}

func (s *Mount) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	if s.finalMountPath== "" {
		return
	}

	ui.Say("Unmounting the root device...")
	umountCommand := fmt.Sprintf("sudo umount %s", s.finalMountPath)
	if _, err := communicator.RunCommand(umountCommand, state, nil); err != nil {
		return
	}

	s.finalMountPath = ""
	return
}
