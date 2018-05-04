
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
	"io/ioutil"
	"fmt"

	"github.com/hashicorp/hcl"
)

const (
	CONFIG_MEL_FILE = "config.mel"
)

func main() {
	var cfg_data []byte
	var err error

	// get contents from config.mel
	if cfg_data, err = ioutil.ReadFile(CONFIG_MEL_FILE); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// check if the contents has HCL syntax
	if _, err = hcl.Parse(string(cfg_data)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create go file with config.mel
	cfg_fp, err := os.Create("config_mel.go")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg_fp.WriteString("package main\n\n")
	cfg_fp.WriteString("const (\nCONFIG_HCL_OPTS=`")
	cfg_fp.WriteString(string(cfg_data))
	cfg_fp.WriteString("`)\n")

	cfg_fp.Close()

	os.Exit(0)
}
