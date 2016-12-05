package builder

import (
	"errors"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ImageName         string     `mapstructure:"image_name"`
	ImageSize         string     `mapstructure:"image_size"`
	ChrootMounts      [][]string `mapstructure:"chroot_mounts"`
	CommandWrapper    string     `mapstructure:"command_wrapper"`
	CopyFiles         []string   `mapstructure:"copy_files"`
	MountOptions      []string   `mapstructure:"mount_options"`
	MountPartition    int        `mapstructure:"mount_partition"`
	MountPath         string     `mapstructure:"mount_path"`
	PostMountCommands []string   `mapstructure:"post_mount_commands"`
	PreMountCommands  []string   `mapstructure:"pre_mount_commands"`

	Ctx               interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	var c Config

	err := config.Decode(&c, &config.DecodeOpts{
		Interpolate:       true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"command_wrapper",
				"post_mount_commands",
				"pre_mount_commands",
				"mount_path",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Defaults
	if c.ChrootMounts == nil {
		c.ChrootMounts = make([][]string, 0)
	}

	if c.CopyFiles == nil {
		c.CopyFiles = make([]string, 0)
	}

	if len(c.ChrootMounts) == 0 {
		c.ChrootMounts = [][]string{
			{"proc", "proc", "/proc"},
			{"sysfs", "sysfs", "/sys"},
			{"bind", "/dev", "/dev"},
			{"devpts", "devpts", "/dev/pts"},
			{"binfmt_misc", "binfmt_misc", "/proc/sys/fs/binfmt_misc"},
		}
	}

	if len(c.CopyFiles) == 0 {
		c.CopyFiles = []string{"/etc/resolv.conf"}
	}

	if c.CommandWrapper == "" {
		c.CommandWrapper = "{{.Command}}"
	}

	if c.MountPath == "" {
		c.MountPath = "packer-chroot-volumes/{{.Device}}"
	}

	if c.MountPartition == 0 {
		c.MountPartition = 1
	}

	var errs *packer.MultiError
	var warns []string

	for _, mounts := range c.ChrootMounts {
		if len(mounts) != 3 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("Each chroot_mounts entry should be three elements."))
			break
		}
	}

	if len(c.PreMountCommands) == 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("pre_mount_commands is required."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warns, errs
	}

	return &c, warns, nil
}
