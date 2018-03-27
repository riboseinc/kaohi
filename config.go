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
 	"io/ioutil"
 	"sync"

 	"github.com/bitly/go-simplejson"
)

// kaohi configuration structure
type kActionConfig struct {
	name string
	actType string

	// file paths for tail mode
	file_paths []string

	// commands for command mode
	cmds []string
}

type kConfig struct {
	log_dir string
	log_level string

	listen_addr string

	actConfigs []kActionConfig

	mux sync.Mutex
}

// add config item
func (config *kConfig) addConfigItem(js *simplejson.Json) {
	var actConfig kActionConfig

	// set name and type
	actConfig.name, _ = js.Get("name").String()
	actConfig.actType, _ = js.Get("type").String()

	if actConfig.actType == "tail" {
		file_paths, _ := js.Get("file_paths").Array()
		for _, v := range file_paths {
			actConfig.file_paths = append(actConfig.file_paths, v.(string))
		}
	} else if actConfig.actType == "command" {
		cmds, _ := js.Get("cmds").Array()
		for _, v := range cmds {
			actConfig.cmds = append(actConfig.cmds, v.(string))
		}
	}

	config.actConfigs = append(config.actConfigs, actConfig)
}

// init config
func (config *kConfig) InitConfig(file_path string) error {

	fmt.Println("Initialize configuration from file", file_path)

	// read file contents
	config_jdata, err := ioutil.ReadFile(file_path)
	if err != nil {
		return ErrConfigOpenFailed
	}

	// parse JSON configuration using simeplejson package
	js, err := simplejson.NewJson(config_jdata)
	if err != nil {
		return ErrConfigInvalidJSON
	}

	// get log directory path
	config.log_dir, err = js.Get("log_dir").String()
	if err != nil {
		config.log_dir = KAOHI_DEFAULT_LOG_DIR
	}

	config.log_level, err = js.Get("log_level").String()
	if err != nil {
		config.log_level = KAOHI_DEFAULT_LOG_LEVEL
	}

	// get listen address
	config.listen_addr, err = js.Get("listen_address").String()
	if err != nil {
		config.listen_addr = KAOHI_DEFAULT_LISTEN_ADDR
	}

	// get action configs
	actConfigs, _ := js.Get("configs").Array()
	for i := 0; i < len(actConfigs); i++ {
		config.addConfigItem(js.Get("configs").GetIndex(i))
	}

	return nil
}

// add config item
func (config *kConfig) AddConfigItem(name string, actType string, actData []string) error {
	var configItem kActionConfig

	// set config item
	configItem.name = name
	configItem.actType = actType

	if actType == "tail" {
		configItem.file_paths = actData
	} else if actType == "command" {
		configItem.cmds = actData
	}

	// add config item to list
	config.mux.Lock()

	for i := 0; i < len(config.actConfigs); i++ {
		if name == config.actConfigs[i].name {
			config.mux.Unlock()
			return ErrConfigItemExist
		}
	}

	config.actConfigs = append(config.actConfigs, configItem)

	config.mux.Unlock()

	return nil
}

// remove config item
func (config *kConfig) RemoveConfigItem(name string) error {
	var i int

	config.mux.Lock()

	// find config with same name
	for i := 0; i < len(config.actConfigs); i++ {
		if name == config.actConfigs[i].name {
			break;
		}
	}

	if i == len(config.actConfigs) {
		config.mux.Unlock()
		return ErrConfigNoExist
	}

	// remove config
	config.actConfigs = append(config.actConfigs[:i], config.actConfigs[i+1:]...)

	config.mux.Unlock()

	return nil
}

// save config to file
func (config *kConfig) SaveConfig(file_path string) error {
	var js *simplejson.Json

	config.mux.Lock()

	// marshall JSON structure
	js = simplejson.New()

	configItems := make([]*simplejson.Json, len(config.actConfigs))
	for i := 0; i < len(config.actConfigs); i++ {
		actConfig := config.actConfigs[i]
		jsActConfig := simplejson.New()

		jsActConfig.Set("name", actConfig.name)
		jsActConfig.Set("type", actConfig.actType)

		fmt.Println(actConfig.name, actConfig.actType)

		if actConfig.actType == "tail" {
			jsActConfig.Set("file_paths", actConfig.file_paths)
		} else {
			jsActConfig.Set("cmds", actConfig.cmds)
		}
		configItems[i] = jsActConfig
	}
	js.Set("configs", configItems)

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

// get log directory
func (config *kConfig) GetLogDir() string {
	return config.log_dir
}

// get log level
func (config *kConfig) GetLogLevel() string {
	return config.log_level
}