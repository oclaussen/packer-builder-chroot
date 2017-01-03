package communicator

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/mitchellh/packer/packer"
)

type Shell struct{
	CmdWrapper CommandWrapper
}

func (c *Shell) Start(cmd *packer.RemoteCmd) error {
	var err error
	command := cmd.Command

	command, err = c.CmdWrapper(command)
	if err != nil {
		return fmt.Errorf("Error wrapping command: %s", err)
	}

	localCmd := exec.Command("/bin/sh", "-c", command)
	localCmd.Stdin = cmd.Stdin
	localCmd.Stdout = cmd.Stdout
	localCmd.Stderr = cmd.Stderr

	log.Printf("Executing: %s %#v", localCmd.Path, localCmd.Args)

	if err := localCmd.Start(); err != nil {
		return err
	}

	go func() {
		var exitStatus int
		if err := localCmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitStatus = 1
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					exitStatus = status.ExitStatus()
				}
			}
		}
		cmd.SetExited(exitStatus)
	}()

	return nil
}

func (c *Shell) Upload(string, io.Reader, *os.FileInfo) error {
	return fmt.Errorf("upload not supported")
}

func (c *Shell) UploadDir(string, string, []string) error {
	return fmt.Errorf("uploadDir not supported")
}

func (c *Shell) Download(string, io.Writer) error {
	return fmt.Errorf("download not supported")
}

func (c *Shell) DownloadDir(src string, dst string, exclude []string) error {
	return fmt.Errorf("downloadDir not supported")
}
