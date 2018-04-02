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
	"strconv"
	"strings"
	"path/filepath"
)

// Service constants
const (
	success = "\t\t\t\t\t[  \033[32mOK\033[0m  ]" // Show colored "OK"
	failed  = "\t\t\t\t\t[\033[31mFAILED\033[0m]" // Show colored "FAILED"
)

// Lookup path for executable file
func executablePath(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return filepath.Abs(os.Args[0])
		}
		return path, nil
	}
	return filepath.Abs(os.Args[0])
}

// Check root rights to use system service
func checkPrivileges() (bool, error) {
	if output, err := exec.Command("id", "-g").Output(); err == nil {
		if gid, parseErr := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 32); parseErr == nil {
			if gid == 0 {
				return true, nil
			}
			return false, ErrRootPriveleges
		}
	}
	return false, ErrUnsupportedSystem
}

// fill array with specified character
func FillBytesArray(len int, ch byte) []byte {
	bytes_arr := make([]byte, len)
	for i := range bytes_arr {
		bytes_arr[i] = ch
	}

	return bytes_arr
}
