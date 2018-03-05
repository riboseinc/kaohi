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
	"os/exec"
	"path/filepath"
	"regexp"
	"text/template"
)

// Daemon interface has a standard set of methods/commands
type Daemon interface {

	// Install the service into the system
	Install(args ...string) (string, error)

	// Remove the service and all corresponding files from the system
	Remove() (string, error)

	// Start the service
	Start() (string, error)

	// Stop the service
	Stop() (string, error)

	// Status - check the service status
	Status() (string, error)
}

// ExecPath tries to get executable path
func ExecPath() (string, error) {
	return execPath()
}

/* darwinRecord - standard record (struct) for darwin version of daemon package */
type darwinRecord struct {
	name         string
	description  string
	dependencies []string
}

/* create new instance of daemon */
func newDaemon(name, description string, dependencies []string) (Daemon, error) {
	return &darwinRecord{name, description, dependencies}, nil
}

/* standard service path for system daemons */
func (darwin *darwinRecord) servicePath() string {
	return "/Library/LaunchDaemons/" + darwin.name + ".plist"
}

/* check if the service is already installed */
func (darwin *darwinRecord) isInstalled() bool {
	if _, err := os.Stat(darwin.servicePath()); err == nil {
		return true
	}

	return false
}

/* get executable path */
func execPath() (string, error) {
	return filepath.Abs(os.Args[0])
}

/* check if the service is already running */
func (darwin *darwinRecord) checkRunning() (string, bool) {
	output, err := exec.Command("launchctl", "list", darwin.name).Output()
	if err == nil {
		
		if matched, err := regexp.MatchString(darwin.name, string(output)); err == nil && matched {
			reg := regexp.MustCompile("PID\" = ([0-9]+);")
			data := reg.FindStringSubmatch(string(output))
			if len(data) > 1 {
				return "Service (pid  " + data[1] + ") is running...", true
			}
			return "Service is running...", true
		}
	}

	return "Service is stopped", false
}

/* install new service */
func (darwin *darwinRecord) Install(args ...string) (string, error) {
	installAction := "Install " + darwin.description + ":"

	srvPath := darwin.servicePath()

	if darwin.isInstalled() {
		return installAction + failed, ErrAlreadyInstalled
	}

	file, err := os.Create(srvPath)
	if err != nil {
		return installAction + failed, err
	}
	defer file.Close()

	execPatch, err := executablePath(darwin.name)
	if err != nil {
		return installAction + failed, err
	}

	templ, err := template.New("propertyList").Parse(propertyList)
	if err != nil {
		return installAction + failed, err
	}

	if err := templ.Execute(
		file,
		&struct {
			Name, Path string
			Args       []string
		}{darwin.name, execPatch, args},
	); err != nil {
		return installAction + failed, err
	}

	return installAction + success, nil
}

// Remove the service
func (darwin *darwinRecord) Remove() (string, error) {
	removeAction := "Removing " + darwin.description + ":"

	if !darwin.isInstalled() {
		return removeAction + failed, ErrNotInstalled
	}

	if err := os.Remove(darwin.servicePath()); err != nil {
		return removeAction + failed, err
	}

	return removeAction + success, nil
}

// Start the service
func (darwin *darwinRecord) Start() (string, error) {
	startAction := "Starting " + darwin.description + ":"

	if !darwin.isInstalled() {
		return startAction + failed, ErrNotInstalled
	}

	if _, ok := darwin.checkRunning(); ok {
		return startAction + failed, ErrAlreadyRunning
	}

	if err := exec.Command("launchctl", "load", darwin.servicePath()).Run(); err != nil {
		return startAction + failed, err
	}

	return startAction + success, nil
}

// Stop the service
func (darwin *darwinRecord) Stop() (string, error) {
	stopAction := "Stopping " + darwin.description + ":"

	if !darwin.isInstalled() {
		return stopAction + failed, ErrNotInstalled
	}

	if _, ok := darwin.checkRunning(); !ok {
		return stopAction + failed, ErrAlreadyStopped
	}

	if err := exec.Command("launchctl", "unload", darwin.servicePath()).Run(); err != nil {
		return stopAction + failed, err
	}

	return stopAction + success, nil
}

// Status - Get service status
func (darwin *darwinRecord) Status() (string, error) {

	if !darwin.isInstalled() {
		return "Status could not defined", ErrNotInstalled
	}

	statusAction, _ := darwin.checkRunning()

	return statusAction, nil
}

var propertyList = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>KeepAlive</key>
	<true/>
	<key>Label</key>
	<string>{{.Name}}</string>
	<key>ProgramArguments</key>
	<array>
	    <string>{{.Path}}</string>
		{{range .Args}}<string>{{.}}</string>
		{{end}}
	</array>
	<key>RunAtLoad</key>
	<true/>
    <key>WorkingDirectory</key>
    <string>/usr/local/var</string>
    <key>StandardErrorPath</key>
    <string>/usr/local/var/log/{{.Name}}.err</string>
    <key>StandardOutPath</key>
    <string>/usr/local/var/log/{{.Name}}.log</string>
</dict>
</plist>
`
