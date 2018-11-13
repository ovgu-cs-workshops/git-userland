/* user-image - service to run one or more web terminals
 *
 * Copyright (C) 2018
 *     Martin Koppehel <martin@embedded.enterprises>,
 *
 */

package main

import (
	"context"
	"os"
	"runtime"

	"github.com/ovgu-cs-workshops/git-userland/tui"
	"github.com/ovgu-cs-workshops/git-userland/util"

	"github.com/EmbeddedEnterprises/service"
	"github.com/gammazero/nexus/client"
	"github.com/gammazero/nexus/wamp"
)

var instanceid string
var username string

func main() {
	app := service.New(service.Config{
		Name:          "git-userland",
		Serialization: client.MSGPACK,
		Version:       "0.1.0",
		Description:   "Service to manage the userland",
	})
	util.Log = app.Logger
	util.App = app
	if iid, ok := os.LookupEnv("RUNINST"); !ok {
		util.Log.Criticalf("RUNINST env variable is required but not set!")
		os.Exit(service.ExitArgument)
	} else {
		instanceid = iid
	}
	if user, ok := os.LookupEnv("RUNUSER"); !ok {
		util.Log.Criticalf("RUNUSER env variable is required but not set!")
		os.Exit(service.ExitArgument)
	} else {
		username = user
	}

	app.Connect()

	procedures := map[string]service.HandlerRegistration{
		"rocks.git.tui." + instanceid + ".create": service.HandlerRegistration{
			Handler: createTui,
			Options: wamp.Dict{
				wamp.OptDiscloseCaller: true,
			},
		},
	}

	if err := app.RegisterAll(procedures); err != nil {
		util.Log.Errorf("Failed to register procedure: %s", err)
		os.Exit(service.ExitRegistration)
	}

	if err := app.Client.Subscribe(string(wamp.MetaEventSessionOnLeave), func(args wamp.List, _, _ wamp.Dict) {
		go func() {
			sid, ok := wamp.AsID(args[0])
			util.Log.Debugf("session %d left, %v", sid, ok)
			tui.OnSessionLeave(sid)
		}()
	}, nil); err != nil {
		util.Log.Errorf("Failed to subscribe metaevent: %s", err)
		os.Exit(service.ExitRegistration)
	}

	app.Run()

	util.Log.Infof("After main, %d goroutines running", runtime.NumGoroutine())
	tui.Shutdown()
	os.Exit(service.ExitSuccess)
}

func createTui(_ context.Context, args wamp.List, _, details wamp.Dict) *client.InvokeResult {
	if len(args) < 3 {
		return service.ReturnError("rocks.git.invalid-argument")
	}
	cid, ok := wamp.AsString(args[0])
	width, wok := wamp.AsID(args[1])
	height, hok := wamp.AsID(args[2])
	callerid, idok := wamp.AsID(details["caller"])
	calleruser, userok := wamp.AsString(details["caller_authid"])
	if !ok || !wok || !hok || !idok || !userok || width > 0xffff || height > 0xffff {
		return service.ReturnError("rocks.git.invalid-argument")
	}
	if username != calleruser {
		return service.ReturnError("rocks.git.not-authorized")
	}
	util.Log.Debugf("Running tui for caller: %v", callerid)
	if err := tui.RunNew(instanceid, cid, uint16(width), uint16(height), callerid); err != nil {
		util.Log.Warningf("Failed to run instance: %v", err)
		return service.ReturnError("rocks.git.internal-error")
	}
	return service.ReturnEmpty()
}
