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

// kaohi context structure
type kContext struct {
	config *kConfigScheme
	logger *kLogger
}

func NewKaohiContext() *kContext {
	return &kContext {
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

// loop until interupt has occurre
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
	var err error

	// create new context
	ctx := NewKaohiContext()

	// check sudo privilege
	if ok, err := checkPrivileges(); !ok {
		fmt.Println(err)
		os.Exit(1)
	}

	// parse configuration file
	if err = ctx.config.ParseConfig(KAOHI_DEFAULT_CONFIG_FILE); err != nil {
		fmt.Println(err)
		os.Exit(1)
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
