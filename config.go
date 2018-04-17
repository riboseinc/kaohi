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
	"os"
	"fmt"
	"io/ioutil"
	"sync"
	"net"
	"strconv"
	"errors"

	"github.com/riboseinc/go-multiconfig"
)

var ConfigOpts = []mconfig.ConfigItemScheme {
	{
		"config",
		"The configuration file path",
		CONF_VAL_TYPE_STRING,
		&CommandLineSwitch{'c', "config"},
		"", "",
		"Specify the configuration file path",
		nil,
	},

	{
		"log-dir",
		"The logging directory",
		CONF_VAL_TYPE_STRING,
		&CommandLineSwitch{'d', "log-dir"},
		"log-settings.directory", "",
		"Specify the logging directory",
		nil,
	},

	{
		"listen",
		"Listening address for incoming events",
		CONF_VAL_TYPE_IPADDR,
		&CommandLineSwitch{'l', "listen"},
		"listen.address", "",
		"Specify the listenning address",
		nil,
	},
}


// kaohi configuration structure
type kTailItem struct {
	name       string
	file_paths []string
}

type kCmdItem struct {
	name     string
	uid      int
	interval int
	cmds     []string
}

type kRsyslogConfig struct {
	status          bool
	listen_address  string
	protocol        string
	local_sock_path string
}

func NewRsyslogConfig() *kRsyslogConfig {
	return &kRsyslogConfig {
		status:           false,
		listen_address:   KAOHI_DEFAULT_SYSLOG_LISTEN_ADDR,
		protocol:         KAOHI_DEFAULT_SYSLOG_PROTO,
		local_sock_path:  "",
	}
}

type kConfig struct {
	log_dir        string
	log_level      string

	listen_addr    string

	rsyslog_config *kRsyslogConfig

	tail_items     []kTailItem
	cmd_items      []kCmdItem

	mux sync.Mutex
}

func NewKaohiConfig() *kConfig {
	return &kConfig {
		log_dir:         KAOHI_DEFAULT_LOG_DIR,
		log_level:       KAOHI_DEFAULT_LOG_LEVEL,

		listen_addr:     KAOHI_DEFAULT_LISTEN_ADDR,

		rsyslog_config:  NewRsyslogConfig(),

		mux:             sync.Mutex{},
	}
}

// add config item
func (config *kConfig) AddTailItem(name string, file_paths []string) error {
	var tail_item kTailItem

	// set tail item
	tail_item.name = name
	tail_item.file_paths = file_paths

	// add config item to list
	config.mux.Lock()

	for i := 0; i < len(config.tail_items); i++ {
		if name == config.tail_items[i].name {
			config.mux.Unlock()
			return ErrConfigItemExist
		}
	}

	config.tail_items = append(config.tail_items, tail_item)

	config.mux.Unlock()

	return nil
}

// add command
func (config *kConfig) AddCmdItem(name string, uid int, interval int, cmds []string) error {
	var cmd_item kCmdItem

	// set cmd item
	cmd_item.name = name
	cmd_item.uid = uid
	cmd_item.interval = interval
	cmd_item.cmds = cmds

	// add cmd item to list
	config.mux.Lock()

	for i := 0; i < len(config.cmd_items); i++ {
		if name == config.cmd_items[i].name {
			config.mux.Unlock()
			return ErrConfigItemExist
		}
	}

	config.cmd_items = append(config.cmd_items, cmd_item)

	config.mux.Unlock()

	return nil
}

// remove tail item
func (config *kConfig) RemoveTailItem(name string) error {
	var i int

	config.mux.Lock()

	// find config with same name
	for i := 0; i < len(config.tail_items); i++ {
		if name == config.tail_items[i].name {
			break;
		}
	}

	if i == len(config.tail_items) {
		config.mux.Unlock()
		return ErrConfigNoExist
	}

	// remove config
	config.tail_items = append(config.tail_items[:i], config.tail_items[i+1:]...)

	config.mux.Unlock()

	return nil
}

// remove cmd item
func (config *kConfig) RemoveCmdItem(name string) error {
	var i int

	config.mux.Lock()

	// find config with same name
	for i := 0; i < len(config.cmd_items); i++ {
		if name == config.cmd_items[i].name {
			break;
		}
	}

	if i == len(config.cmd_items) {
		config.mux.Unlock()
		return ErrConfigNoExist
	}

	// remove config
	config.cmd_items = append(config.cmd_items[:i], config.cmd_items[i+1:]...)

	config.mux.Unlock()

	return nil
}

// save config to file
func (config *kConfig) SaveConfig(file_path string) error {
	var js *simplejson.Json

	config.mux.Lock()

	// marshall JSON structure
	js = simplejson.New()

	js_tail_items := make([]*simplejson.Json, len(config.tail_items))
	for i := 0; i < len(config.tail_items); i++ {
		tail_item := config.tail_items[i]
		js_tail := simplejson.New()

		js_tail.Set("name", tail_item.name)
		js_tail.Set("file_paths", tail_item.file_paths)

		js_tail_items[i] = js_tail
	}
	js.Set("config_files", js_tail_items)

	js_cmd_items := make([]*simplejson.Json, len(config.cmd_items))
	for i := 0; i < len(config.cmd_items); i++ {
		cmd_item := config.cmd_items[i]
		js_cmd := simplejson.New()

		js_cmd.Set("name", cmd_item.name)

		if cmd_item.uid > 0 {
			js_cmd.Set("uid", cmd_item.uid)
		}

		if cmd_item.interval > 0 {
			js_cmd.Set("interval", cmd_item.interval)
		}

		js_cmd.Set("cmds", cmd_item.cmds)

		js_cmd_items[i] = js_cmd
	}
	js.Set("commands", js_cmd_items)

	js.Set("log_dir", config.log_dir)
	js.Set("log_level", config.log_level)

	config.mux.Unlock()

	config_jdata, _ := js.MarshalJSON()
	err := ioutil.WriteFile(file_path, config_jdata, 0644)
	if err != nil {
		return ErrConfigSaveFailed
	}

	return nil
}

// add tail item
func (config *kConfig) addTailItem(js *simplejson.Json) error {
	var tail_item kTailItem
	var err error
	var file_paths []interface{}

	// get name
	if tail_item.name, err = js.Get("name").String(); err != nil {
		return err
	}

	// get file paths
	if file_paths, err = js.Get("file_paths").Array(); err != nil {
		return err
	}

	for _, v := range file_paths {
		tail_item.file_paths = append(tail_item.file_paths, v.(string))
	}

	// add tail item to config
	config.tail_items = append(config.tail_items, tail_item)

	return nil
}

// add command item
func (config *kConfig) addCmdItem(js *simplejson.Json) error {
	var cmd_item kCmdItem
	var err error
	var cmds []interface{}

	// get name
	if cmd_item.name, err = js.Get("name").String(); err != nil {
		return err
	}

	if cmd_item.uid, err = js.Get("uid").Int(); err != nil {
		cmd_item.uid = KAOHI_DEFAULT_CMD_UID
	}

	if cmd_item.interval, err = js.Get("uid").Int(); err != nil {
		cmd_item.interval = KAOHI_DEFAULT_CMD_INTERVAL
	}

	// get cmd list
	if cmds, err = js.Get("cmds").Array(); err != nil {
		return err
	}

	for _, v := range cmds {
		cmd_item.cmds = append(cmd_item.cmds, v.(string))
	}

	// add command item to config
	config.cmd_items = append(config.cmd_items, cmd_item)

	return nil
}

// parse rsyslog config
func (config *kConfig) parseRsyslogConfig(rsyslog_js *simplejson.Json) error {
	var status bool
	var listen_addr, protocol, local_sock_path string
	var err error

	// get rsyslog status
	if status, err = rsyslog_js.Get("status").Bool(); err != nil {
		config.rsyslog_config.status = false
	}
	config.rsyslog_config.status = status

	// get listen address
	if listen_addr, err = rsyslog_js.Get("listen_addr").String(); err != nil {
		config.rsyslog_config.listen_address = ""
	}
	config.rsyslog_config.listen_address = listen_addr;

	// get protocol
	if protocol, err = rsyslog_js.Get("protocol").String(); err == nil {
		config.rsyslog_config.protocol = protocol
	}

	// get local socket path
	if local_sock_path, err = rsyslog_js.Get("local_sock_path").String(); err == nil {
		config.rsyslog_config.local_sock_path = local_sock_path
	}

	return nil
}

// get log directory
func (config *kConfig) GetLogDir() string {
	return config.log_dir
}

// get log level
func (config *kConfig) GetLogLevel() string {
	return config.log_level
}

// init config
func InitConfig(file_path string) (*kConfig, error) {
	var config_jdata []byte
	var js *simplejson.Json
	var log_dir, log_level, listen_addr string
	var err error

	fmt.Println("Initialize configuration from file", file_path)

	// create new config
	config := NewKaohiConfig();

	// read file contents
	if config_jdata, err = ioutil.ReadFile(file_path); err != nil {
		return nil, ErrConfigOpenFailed
	}

	// parse JSON configuration using simeplejson package
	if js, err = simplejson.NewJson(config_jdata); err != nil {
		return nil, ErrConfigInvalidJSON
	}

	// get log directory path
	if log_dir, err = js.Get("log_dir").String(); err == nil {
		config.log_dir = log_dir
	}

	if log_level, err = js.Get("log_level").String(); err == nil {
		config.log_level = log_level
	}

	// get listen address
	if listen_addr, err = js.Get("listen_address").String(); err == nil {
		config.listen_addr = listen_addr
	}

	// get tail items
	if tail_items, err := js.Get("config_files").Array(); err == nil {
		for i := 0; i < len(tail_items); i++ {
			config.addTailItem(js.Get("config_files").GetIndex(i))
		}
	}

	// get command items
	if cmd_items, err := js.Get("commands").Array(); err == nil {
		for i := 0; i < len(cmd_items); i++ {
			config.addCmdItem(js.Get("commands").GetIndex(i))
		}
	}

	// parse rsyslog info
	if rsyslog_js := js.Get("rsyslog"); rsyslog_js != nil {
		config.parseRsyslogConfig(rsyslog_js)
	}

	return config, nil
}
