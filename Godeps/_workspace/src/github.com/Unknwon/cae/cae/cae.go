// Copyright 2014 Unknown
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// cae is a command-line tool for operating ZIP/TAR.GZ files based on cae package.
package main

import (
	"log"
	"os"

	"github.com/Unknwon/cae/tz"
)

func main() {
	// cae.Copy("test.zip", "zip/testdata/test.zip")
	fw, err := os.Create("hello.tar.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()

	fr, err := os.Open("zip/testdata/gophercolor16x16.png")
	if err != nil {
		log.Fatal(err)
	}

	fi, err := fr.Stat()
	if err != nil {
		log.Fatal(err)
	}

	s := tz.NewStreamArachive(fw)
	if err = s.StreamReader("", fi, fr); err != nil {
		log.Fatal(err)
	}
	if err = s.Close(); err != nil {
		log.Fatal(err)
	}
}
