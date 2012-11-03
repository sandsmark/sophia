// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"

    // Connection stuff
	"code.google.com/p/go.crypto/ssh/terminal"
	"code.google.com/p/go.crypto/ssh"

    // DB stuff
    //"database/sql"
    //_ "github.com/jbarham/gopgsqldriver"

    // Interface stuff
    //"github.com/nsf/termbox-go"

    // Password storage
    //"github.com/jameskeane/bcrypt"
)

func main() {
	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn *ssh.ServerConn, user, pass string) bool {
			return user == "testuser" && pass == "tiger"
		},
	}

	pemBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		panic("Failed to load private key")
	}
	if err = config.SetRSAPrivateKey(pemBytes); err != nil {
		panic("Failed to parse private key")
	}

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := ssh.Listen("tcp", "0.0.0.0:2022", config)
	if err != nil {
		panic("failed to listen for connection")
	}
	sConn, err := listener.Accept()
	if err != nil {
		panic("failed to accept incoming connection")
	}
	if err := sConn.Handshake(); err != nil {
		panic("failed to handshake")
	}

	// A ServerConn multiplexes several channels, which must 
	// themselves be Accepted.
	for {
		// Accept reads from the connection, demultiplexes packets
		// to their corresponding channels and returns when a new
		// channel request is seen. Some goroutine must always be
		// calling Accept; otherwise no messages will be forwarded
		// to the channels.
		channel, err := sConn.Accept()
		if err != nil {
			panic("error from Accept")
		}

		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if channel.ChannelType() != "session" {
			channel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel.Accept()

		term := terminal.NewTerminal(channel, "> ")
		serverTerm := &ssh.ServerTerminal{
			Term:    term,
			Channel: channel,
		}
		go func() {
			defer channel.Close()
			for {
				line, err := serverTerm.ReadLine()
				if err != nil {
					break
				}
				fmt.Println(line)
			}
		}()
	}
}

