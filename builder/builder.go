package builder

import (
	"errors"
	"runtime"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"github.com/oclaussen/packer-builder-chroot/steps"
	"github.com/oclaussen/packer-builder-chroot/communicator"
)

const BuilderId = "oclaussen.chroot"

type Builder struct {
	config *Config
	runner multistep.Runner
}

type wrappedCommandTemplate struct {
	Command string
}
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("The chroot builder only works on Linux environments.")
	}
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c
	return warnings, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {

	wrappedCommand := func(command string) (string, error) {
		ctx := b.config.Ctx
		ctx.Data = &wrappedCommandTemplate{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &ctx)
	}

	// The state
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", communicator.CommandWrapper(wrappedCommand))

	// The steps
	steps := []multistep.Step{
		&steps.Prepare{
			ImageName: b.config.ImageName,
			ImageSize: b.config.ImageSize,
		},
		&steps.PreMount{
			Commands: b.config.PreMountCommands,
		},
		&steps.Mount{
			MountOptions:   b.config.MountOptions,
			MountPartition: b.config.MountPartition,
			MountPath:      b.config.MountPath,
		},
		&steps.PostMount{
			Commands: b.config.PostMountCommands,
		},
		&steps.MountExtra{
			ChrootMounts: b.config.ChrootMounts,
		},
		&steps.CopyFiles{
			Files: make([]string, 0, len(b.config.CopyFiles)),
		},
		&steps.Provision{},
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// No errors, must've worked
	artifact := &NullArtifact{}
	return artifact, nil
}

func (b *Builder) Cancel() {
	b.runner.Cancel()
}
