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
	KAOHI_DEFAULT_CONFIG_FILE        = "/etc/kaohi.conf"

	KAOHI_DEFAULT_LOG_DIR            = "/var/log/kaohi"
	KAOHI_DEFAULT_LOG_LEVEL          = "NORMAL"

	KAOHI_DEFAULT_LISTEN_ADDR        = "127.0.0.1:6688"

	KAOHI_RA_SOCK_PATH               = "/var/run/.kaohi_ra"

	KAOHI_DEFAULT_CMD_UID            = 0
	KAOHI_DEFAULT_CMD_INTERVAL       = -1

	KAOHI_DEFAULT_SYSLOG_STATUE      = 0
	KAOHI_DEFAULT_SYSLOG_LISTEN_ADDR = "0.0.0.0:5443"
	KAOHI_DEFAULT_SYSLOG_PROTO       = "tcp"
)

const (
	KAOHI_HCL_OPTIONS = `cmdline "log_directory" {
			type = "string"
			switch {
				short = "l"
				long = "log-dir"
			}

			description {
				short = "log directory"
				long = "Specify the path of log directory"
			}

			env = "KAOHI_LOG_DIR"
			config = "global.log_directory"
		}

		cmdline "log_level" {
			type = "int"
			switch {
				short = "d"
				long = "verbose"
			}

			description {
				short = "verbose level (0 ~ 3)"
				long = "Specify the verbose level"
			}

			env = "KAOHI_LOG_LEVEL"
			config = "global.log_level"
		}

		cmdline "listen_address" {
			type = "ipport"
			switch {
				short = "L"
				long = "listen"
			}

			description = {
				short = "IP:port"
				long = "Specify the listening address for Kaohi console"
			}

			env = "KAOHI_LISTEN_ADDR"
			config = "global.listen_address"
		}

		cmdline "cmd_interval" {
			type = "int"
			switch {
				short = "i"
				long = "interval"
			}

			description = {
				short = "interval"
				long = "Specify the interval for command execution"
			}

			env = "KAOHI_CMD_INTERVAL"
			config = "commands.*.interval"
		}

		cmdline "cmd_uid" {
			type = "int"
			switch {
				short = "u"
				long = "user-id"
			}

			description = {
				short = "user ID"
				long = "Specify User ID for command execution"
			}

			env = "KAOHI_CMD_UID"
			config = "commands.*.uid"
		}

		cmdline "cmd_files" {
			type = "array"
			switch {
				short = "e"
				long = "commands"
			}

			description = {
				short = "commands"
				long = "The list of commands to be executed"
			}

			env = "KAOHI_CMD_FILES"
			config = "commands.*.files"
		}

		cmdline "rsyslog_listen" {
			type = "ipport"
			switch {
				short = "r"
				long = "rsyslog-listen"
			}

			description = {
				short = "IP:port"
				long = "Specify IP:port for rsyslog listening"
			}

			env = "KAOHI_RSYSLOG_LISTEN"
			config = "rsyslog.listen_address"
		}

		cmdline "rsyslog_proto" {
			type = "proto"
			switch {
				short = "p"
				long = "rsyslog-proto"
			}

			description = {
				short = "tcp|udp"
				long = "Specify the protocol for rsyslog listening"
			}

			env = "KAOHI_RSYSLOG_PROTO"
			config = "rsyslog.protocol"
		}

		cmdline "help" {
			type = "bool"
			switch {
				short = "h"
				long = "help"
			}

			description = {
				short = ""
				long = "Print help message"
			}

			helper = true
		}

		cmdline "config" {
			type = "string"
			switch {
				short = "c"
				long = "config"
			}

			description = {
				short = "config file"
				long = "Specify the configuration file"
			}

			env = "KAOHI_CONFIG"
			override_cfg = true
		}`
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
