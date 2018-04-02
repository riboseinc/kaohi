/*
 * Copyright (c) 2017, [Ribose Inc](https://www.ribose.com).
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var opts = []kCmdLineOptions {
	{'c', "config",        OPT_TYPE_STRING,   true,  "configuration file",       "the path of configuration file"},
	{'l', "listen",        OPT_TYPE_ADDRPAIR, true,  "IP:port",                  "IP address and port number for listen"},
	{'o', "commands",      OPT_TYPE_ARRAY,    true,  "commands",                 "the list of commands to be executed"},
	{'i', "interval",      OPT_TYPE_INT,      true,  "interval",                 "the interval of command execution"},
	{'u', "uid",           OPT_TYPE_INT,      true,  "uid",                      "the UID to run commands"},
	{'w', "config-files",  OPT_TYPE_ARRAY,    true,  "config-files",             "the list of files to be watched"},
	{'d', "log-dir",       OPT_TYPE_STRING,   true,  "log directory path",       "the path of log directory"},
	{'v', "verbose",       OPT_TYPE_INT,      true,  "log level",                "the log level"},
	{'r', "rsyslog",       OPT_TYPE_ADDRPAIR, true,  "rsyslog listening address","IP address and port number for listening rsyslog connections"},
	{'s', "rasock",        OPT_TYPE_STRING,   true,  "reagent socket path",      "the path of unix socket to deliver the events to Reagent"},
	{'p', "cred",          OPT_TYPE_STRING,   true,  "reagent credential",       "the credential to allow Kaohi to authenticate to Reagent"},
	{'k', "key",           OPT_TYPE_STRING,   true,  "reagent key",              "initialization key provided by Reagent to allow signing Kaohi converted events"},
	{'h', "help",          OPT_TYPE_NONE,     false, "",                         "show help message"},
	{'V', "version",       OPT_TYPE_NONE,     false, "",                         "show version number"},
}

// kaohi context structure
type kContext struct {
	options map[string]interface{}

	config *kConfig
	logger *kLogger
}

func NewKaohiContext() *kContext {
	return &kContext {
		options:            make(map[string]interface{}),
		config:             NewKaohiConfig(),
		logger:             nil,
	}
}

// print version
func PrintVersion() {
	
}

// init kaohi context
func (ctx *kContext) Init() error {
	var err error

	// init config
	if ctx.config, err = InitConfig(KAOHI_DEFAULT_CONFIG_FILE); err != nil {
		return err
	}

	// init logging
	if err = InitLogger(ctx.config.GetLogDir(), ctx.config.GetLogLevel()); err != nil {
		return err
	}

	DEBUG_INFO("Initializing Kaohi context")

	// init command listener
	if err = InitCmdListener(ctx); err != nil {
		return err
	}

	// init watcher
	if err = InitKaohiWatcher(); err != nil {
		return err
	}

	return nil
}

// finalize kaohi context
func (ctx *kContext) Finalize() {
	DEBUG_INFO("Finalizing Kaohi context")

	// finalize kaohi watcher
	FinalizeKaohiWatcher()

	// finalize command listener
	FinalizeCmdListener()
}

// loop until interupt has occurred
func WaitForSignal() {
	// create signal channel
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// wait until TERM signal has received
	for {
		select {
		case killSignal := <-interrupt:
			if killSignal == os.Interrupt {
				DEBUG_INFO("Interrupt has occurred by system signal")
				return
			}

			DEBUG_INFO("Kill signal has occurred")
			return
		}
	}
}

// main function
func main() {
	var ctx *kContext
	var err error

	// check sudo privilege
	if ok, err := checkPrivileges(); !ok {
		fmt.Println(err)
		os.Exit(1)
	}

	// parse command line
	if len(os.Args) > 1 {
		if ctx.options, err = ParseCmdLine(opts); err != nil {
			PrintUsage(opts)
			os.Exit(1)
		}

		// print help
		if ctx.options["help"].(bool) == true {
			PrintUsage(opts)
			os.Exit(0)
		}

		// print version
		if ctx.options["version"].(bool) == true {
			PrintVersion()
			os.Exit(0)
		}
	}

	// init context
	if err := ctx.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// main loop
	WaitForSignal()

	// finalize context
	ctx.Finalize()
}
