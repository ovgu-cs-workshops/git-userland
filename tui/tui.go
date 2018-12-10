package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/EmbeddedEnterprises/service"
	"github.com/gammazero/nexus/client"
	"github.com/gammazero/nexus/transport/serialize"
	"github.com/gammazero/nexus/wamp"
	"github.com/kr/pty"
	"github.com/ovgu-cs-workshops/git-userland/util"
)

type processBag struct {
	width, height uint16
	id            string
	instance      string
	ptmx          *os.File
	cmd           *exec.Cmd
	exited        bool
	caller        wamp.ID
	done          chan struct{}
}

var processes map[string]*processBag
var procLock sync.Mutex
var shell string

func init() {
	procLock = sync.Mutex{}
	processes = map[string]*processBag{}
	if sh, ok := os.LookupEnv("USERSHELL"); ok {
		shell = sh
	} else {
		shell = "/bin/zsh"
	}
}

func Shutdown() {
	for _, process := range processes {
		process.kill()
	}
}

func OnSessionLeave(sid wamp.ID) {
	for _, p := range processes {
		p.onSessionLeave(sid)
	}
}

func RunNew(instance, id string, width uint16, height uint16, caller wamp.ID) error {
	procLock.Lock()
	defer procLock.Unlock()
	util.Log.Debugf("Starting tui %s.%s with w=%d, h=%d for caller %d", instance, id, width, height, caller)
	if _, ok := processes[id]; ok {
		return errors.New("already exists")
	}
	c := exec.Command(shell)
	c.Env = []string{
		"TERM=xterm",
	}
	c.Dir = "/home/user"
	//c.SysProcAttr = &syscall.SysProcAttr{}
	//c.SysProcAttr.Credential = &syscall.Credential{Uid: 1000, Gid: 1000}
	ptmx, err := pty.StartWithSize(c, &pty.Winsize{
		Rows: height,
		Cols: width,
	})
	if err != nil {
		util.Log.Debugf("Failed to start pty: %v", err)
		return err
	}
	bag := &processBag{
		height:   height,
		width:    width,
		id:       id,
		instance: instance,
		ptmx:     ptmx,
		cmd:      c,
		exited:   false,
		caller:   caller,
		done:     make(chan struct{}),
	}
	go bag.monitorExit()
	go bag.monitorOutput()
	util.Log.Debugf("Command and monitors running, registering")
	if err := util.App.Client.Register(fmt.Sprintf("rocks.git.tui.%s.%s.input", instance, id), bag.sendInput, wamp.Dict{
		wamp.OptDiscloseCaller: true,
	}); err != nil {
		util.Log.Warningf("Failed to register input proc for session %s: %v", id, err)
	}
	if err := util.App.Client.Register(fmt.Sprintf("rocks.git.tui.%s.%s.resize", instance, id), bag.resize, wamp.Dict{
		wamp.OptDiscloseCaller: true,
	}); err != nil {
		util.Log.Warningf("Failed to register resize proc for session %s: %v", id, err)
	}
	processes[id] = bag
	util.Log.Debugf("Got it!")
	return nil
}

func (p *processBag) monitorOutput() {
	data := make([]byte, 1024)
	topic := fmt.Sprintf("rocks.git.tui.%s.%s.out", p.instance, p.id)
	util.Log.Debugf("Publishing output to '%s'", topic)
	for {
		n, err := p.ptmx.Read(data)
		if err != nil {
			util.Log.Warningf("Failed to read the stream: %v", err)

			break
		}

		pub := util.App.Client.Publish(topic, wamp.Dict{
			wamp.OptAcknowledge: true,
			wamp.WhitelistKey:   []wamp.ID{p.caller},
		}, wamp.List{
			serialize.BinaryData(data[:n]),
		}, nil)
		if pub != nil {
			util.Log.Warningf("Failed to publish output: %v", pub)
		}
	}
}

func (p *processBag) monitorExit() {
	err := p.cmd.Wait()
	procLock.Lock()
	close(p.done)
	defer procLock.Unlock()
	p.exited = true
	util.Log.Debugf("p.cmd.Wait(%s): %v", p.id, err)
	util.App.Client.Publish(fmt.Sprintf("rocks.git.tui.%s.%s.exit", p.instance, p.id), wamp.Dict{
		wamp.WhitelistKey: wamp.List{p.caller},
	}, nil, nil)

	util.App.Client.Unregister(fmt.Sprintf("rocks.git.tui.%s.%s.input", p.instance, p.id))
	util.App.Client.Unregister(fmt.Sprintf("rocks.git.tui.%s.%s.resize", p.instance, p.id))

	util.Log.Debugf("Remove process %s from list", p.id)
	if err := p.ptmx.Close(); err != nil {
		util.Log.Warningf("Failed to close ptmx for process %s: %v", p.id, err)
	}
	delete(processes, p.id)
}

func (p *processBag) sendInput(_ context.Context, args wamp.List, _, details wamp.Dict) *client.InvokeResult {
	if p.exited {
		return service.ReturnError("rocks.git.no-such-tui")
	}
	caller, cok := wamp.AsID(details["caller"])
	if len(args) < 1 || !cok {
		return service.ReturnError("rocks.git.invalid-argument")
	}
	data, ok := args[0].(serialize.BinaryData)
	if !ok {
		return service.ReturnError("rocks.git.invalid-argument")
	}
	if caller != p.caller {
		return service.ReturnError("rocks.git.not-authorized")
	}
	if n, err := p.ptmx.Write([]byte(data)); err != nil || n < len(data) {
		return service.ReturnError("rocks.git.internal-error")
	}
	return service.ReturnEmpty()
}

func (p *processBag) resize(_ context.Context, args wamp.List, _, details wamp.Dict) *client.InvokeResult {
	if p.exited {
		return service.ReturnError("rocks.git.no-such-tui")
	}
	caller, cok := wamp.AsID(details["caller"])
	if len(args) < 2 || !cok {
		return service.ReturnError("rocks.git.invalid-argument")
	}
	w, wok := wamp.AsID(args[0])
	h, hok := wamp.AsID(args[1])
	if !wok || !hok || w > 0xffff || h > 0xffff {
		return service.ReturnError("rocks.git.invalid-argument")
	}
	if caller != p.caller {
		return service.ReturnError("rocks.git.not-authorized")
	}
	if err := pty.Setsize(p.ptmx, &pty.Winsize{
		Rows: uint16(h),
		Cols: uint16(w),
	}); err != nil {
		util.Log.Warningf("Failed to resize pty: %v", err)
		return service.ReturnError("rocks.git.internal-error")
	}
	return service.ReturnEmpty()
}

func (p *processBag) onSessionLeave(sid wamp.ID) {
	if sid == p.caller {
		util.Log.Debugf("Killing session %s for caller %d due to session close", p.id, p.caller)
		p.kill()
	}
}

func (p *processBag) kill() {
	if p.exited {
		return
	}
	util.Log.Debugf("Sending sigint to %s", p.id)
	if err := syscall.Kill(p.cmd.Process.Pid, syscall.SIGINT); err != nil {
		util.Log.Warningf("Failed to int session %s", p.id)
	}
	select {
	case <-p.done:
		return
	case <-time.After(1 * time.Second):
	}
	util.Log.Debugf("Sending sigterm to %s", p.id)
	if err := syscall.Kill(p.cmd.Process.Pid, syscall.SIGTERM); err != nil {
		util.Log.Warningf("Failed to term session %s", p.id)
	}
	select {
	case <-p.done:
		return
	case <-time.After(1 * time.Second):
	}
	util.Log.Debugf("Sending sigquit to %s", p.id)
	if err := syscall.Kill(p.cmd.Process.Pid, syscall.SIGQUIT); err != nil {
		util.Log.Warningf("Failed to quit session %s", p.id)
	}
	select {
	case <-p.done:
		return
	case <-time.After(1 * time.Second):
	}
	util.Log.Debugf("Sending sigkill to %s", p.id)
	if err := syscall.Kill(p.cmd.Process.Pid, syscall.SIGKILL); err != nil {
		util.Log.Warningf("Failed to kill session %s", p.id)
	}
	select {
	case <-p.done:
		return
	case <-time.After(1 * time.Second):
	}
	util.Log.Errorf("Failed to kill tui: %s", p.id)
}
