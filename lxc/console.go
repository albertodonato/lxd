package main

import (
	"os"
	"syscall"

	"github.com/lxc/lxd/client"
	"github.com/lxc/lxd/lxc/config"
	"github.com/lxc/lxd/shared/api"
	"github.com/lxc/lxd/shared/i18n"
	"github.com/lxc/lxd/shared/termios"
)

type consoleCmd struct{}

func (c *consoleCmd) showByDefault() bool {
	return true
}

func (c *consoleCmd) usage() string {
	return i18n.G(
		`Usage: lxc console [<remote>:]<container>

attach to a container console.`)
}

func (c *consoleCmd) flags() {}

func (c *consoleCmd) run(conf *config.Config, args []string) error {
	if len(args) != 1 {
		return errArgs
	}

	remote, name, err := conf.ParseRemote(args[0])
	if err != nil {
		return err
	}

	d, err := conf.GetContainerServer(remote)
	if err != nil {
		return err
	}

	cfd := int(syscall.Stdin)
	oldttystate, err := termios.MakeRaw(cfd)
	if err != nil {
		return err
	}
	defer termios.Restore(cfd, oldttystate)

	width, height, err := termios.GetSize(int(syscall.Stdout))
	if err != nil {
		return err
	}

	req := api.ContainerConsolePost{
		Width:  width,
		Height: height,
	}
	consoleArgs := &lxd.ContainerConsoleArgs{
		Stdin:    os.Stdin,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		DataDone: make(chan bool),
	}
	op, err := d.ConsoleContainer(name, req, consoleArgs)
	if err != nil {
		return err
	}

	// Wait for the operation to complete
	err = op.Wait()
	if err != nil {
		return err
	}

	// Wait for any remaining I/O to be flushed
	<-consoleArgs.DataDone
	os.Exit(0)
	return nil
}
