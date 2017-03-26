// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package nodego provides utilities for pretending to be node.
package nodego

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func startServer(l net.Listener, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		log.Println(http.Serve(l, nil))
		l.Close()
		wg.Done()
	}()
}

const readyMessage = `HTTP/1.0 200 OK
Date: %s
Content-Length: 23
Content-Type: text/plain; charset=utf-8

User function is ready
`

func sendReady(c net.Conn) error {
	_, err := c.Write([]byte(fmt.Sprintf(readyMessage, time.Now().UTC().Format(http.TimeFormat))))
	return err
}

const HTTPTrigger = "/req"

// TakeOver attempts to take over all of node's sockets that were open when it
// execve'd this binary. This binary must have been started by the execer node
// module for this to work.
func TakeOver() {
	ready := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User function is ready")
	}
	http.HandleFunc("/start", ready)
	http.HandleFunc("/check", ready)
	http.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	var wg sync.WaitGroup
	for _, arg := range os.Args[1:] {
		fd, err := strconv.Atoi(arg)
		if err != nil {
			log.Printf("Error converting arg %q to int: %v", arg, err)
			continue
		}
		f := os.NewFile(uintptr(fd), "")
		c, err := net.FileConn(f)
		if err != nil {
			log.Println("Error creating FileConn:", err)
			f.Close()
			continue
		}
		err = sendReady(c)
		c.Close()
		if err == nil {
			f.Close()
			continue
		}
		l, err := net.FileListener(f)
		f.Close()
		if err != nil {
			log.Println("Error creating FileListener:", err)
			continue
		}
		startServer(l, &wg)
	}

	wg.Wait()
}
