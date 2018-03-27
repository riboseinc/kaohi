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
	"errors"
)

// default option values
const (
	KAOHI_DEFAULT_CONFIG_FILE = "/etc/kaohi.conf"

	KAOHI_DEFAULT_LOG_DIR = "/var/log/kaohi"
	KAOHI_DEFAULT_LOG_LEVEL = "NORMAL"

	KAOHI_DEFAULT_LISTEN_ADDR = "127.0.0.1:6688"
)

var (
	ErrUnsupportedSystem = errors.New("Unsupported system")

	ErrRootPriveleges = errors.New("You must have root user privileges. Possibly using 'sudo' command should help")

	ErrAlreadyInstalled = errors.New("Service has already been installed")

	ErrNotInstalled = errors.New("Service is not installed")

	ErrAlreadyRunning = errors.New("Service is already running")

	ErrAlreadyStopped = errors.New("Service has already been stopped")

	ErrInvalidJSON = errors.New("Invalid JSON string")

	// errors related with config
	ErrConfigOpenFailed = errors.New("Could not open configuration file")

	ErrConfigInvalidJSON = errors.New("Invalid JSON configuration file")	

	ErrConfigItemExist = errors.New("The configuration item is already exist with same name")

	ErrConfigNoExist = errors.New("The configuration item isn't exist")

	ErrConfigSaveFailed = errors.New("The configuration file could not be opened for writting")

	// errors related with logger
	ErrCreateLogDir = errors.New("Could not create log directory")

	ErrCreateLogFile = errors.New("Could not create log file")

	// errors related with command listener
	ErrResolveAddr = errors.New("Could not resolve address for listen")

	ErrListenFaield = errors.New("Could not listen on specified address")

	ErrConnClosing   = errors.New("use of closed network connection")

	ErrWriteBlocking = errors.New("write packet was blocking")

	ErrReadBlocking  = errors.New("read packet was blocking")

	// errors related with watcher
	ErrWatchedFileDeleted = errors.New("The wathed file was deleted")
)
