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
	"github.com/riboseinc/go-multiconfig"
)

type kGlobalConfig struct {
	LogDir         string             `hcl:"log_directory"`
	LogLevel       int                `hcl:"log_level"`

	ListenAddr     string             `hcl:"listen_address"`
}

type kFilesConfig struct {
	Name           string             `hcl:",key"`
	Files          []string           `hcl:"files"`
}

type kCommandsConfig struct {
	Name           string             `hcl:",key"`
	Uid            int                `hcl:"uid"`
	Interval       int                `hcl:"interval"`
	Cmds           []string           `hcl:"cmds"`
}

type kRsyslogConfig struct {
	ListenAddr     string             `hcl:"listen_address"`
	Protocol       string             `hcl:"protocol"`
}

type kConfig struct {
	Globals        kGlobalConfig       `hcl:"global"`
	ConfigFiles    []kFilesConfig      `hcl:"config-files"`
	Commands       []kCommandsConfig   `hcl:"commands"`
	Rsyslog        kRsyslogConfig      `hcl:"rsyslog"`
}

type kConfigScheme struct {
	configs        *kConfig
}

func NewKaohiConfig() *kConfigScheme {
	return &kConfigScheme{
		configs:   &kConfig{},
	}
}

func (config *kConfigScheme) ParseConfig(hcl_options string, cfg_path string) error {
	cfg := mconfig.NewConfigScheme()

	if err := cfg.ParseConfig(hcl_options, cfg_path, config.configs); err != nil {
		cfg.PrintCmdLineHelp()
		return err
	}
	fmt.Println("############################Parsing configuration ended")

	return nil
}

func (config *kConfigScheme) GetLogDir() string {
	return config.configs.Globals.LogDir
}

func (config *kConfigScheme) GetLogLevel() int {
	return config.configs.Globals.LogLevel
}

func (config *kConfigScheme) GetListenAddr() string {
	return config.configs.Globals.ListenAddr
}
