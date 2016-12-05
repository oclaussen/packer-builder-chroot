package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/oclaussen/packer-builder-chroot/builder"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(builder.Builder))
	server.Serve()
}
