package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/lxc/lxd/shared"
	"github.com/lxc/lxd/shared/api"
)

type consoleWsOps struct {
	container *containerLXC
}

func (o *consoleWsOps) startup(s *ttyWs) {
}

func (o *consoleWsOps) do(s *ttyWs, stdin, stdout, stderr *os.File) error {
	return nil
}

func (o *consoleWsOps) openTTYs(s *ttyWs) (*os.File, *os.File, *os.File, error) {
	s.ttys = make([]*os.File, 0)
	s.ptys = make([]*os.File, 1)

	fd, err := o.container.c.ConsoleFd(0)
	if err != nil {
		return nil, nil, nil, err
	}
	consoleFile := os.NewFile(uintptr(fd), s.container.Name()+"-console")
	s.ptys[0] = consoleFile

	if s.width > 0 && s.height > 0 {
		shared.SetSize(fd, s.width, s.height)
	}

	return consoleFile, consoleFile, consoleFile, nil
}

func (o *consoleWsOps) getMetadata(s *ttyWs) shared.Jmap {
	return shared.Jmap{}
}

func (o *consoleWsOps) handleSignal(s *ttyWs, signal int) {
}

func (o *consoleWsOps) handleAbnormalClosure(s *ttyWs) {
}

func containerConsolePost(d *Daemon, r *http.Request) Response {
	name := mux.Vars(r)["name"]
	c, err := containerLoadByName(d.State(), name)
	if err != nil {
		return SmartError(err)
	}

	if !c.IsRunning() {
		return BadRequest(fmt.Errorf("Container is not running."))
	}

	if c.IsFrozen() {
		return BadRequest(fmt.Errorf("Container is frozen."))
	}

	post := api.ContainerConsolePost{}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return BadRequest(err)
	}
	if err := json.Unmarshal(buf, &post); err != nil {
		return BadRequest(err)
	}

	// Ensure this is a containerLXC
	cont, ok := c.(*containerLXC)
	if !ok {
		return BadRequest(fmt.Errorf("Operation not supported"))
	}
	ops := &consoleWsOps{container: cont}
	ws, err := newttyWs(ops, c, true, post.Width, post.Height)
	if err != nil {
		return InternalError(err)
	}
	resources := map[string][]string{
		"containers": []string{cont.Name()},
	}
	op, err := operationCreate(operationClassWebsocket, resources, ws.Metadata(), ws.Do, nil, ws.Connect)
	if err != nil {
		return InternalError(err)
	}

	return OperationResponse(op)
}
