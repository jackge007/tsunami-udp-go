package main

// const char* build_time(void) {
//     static const char* t = __DATE__ "  " __TIME__;
//     return t;
// }
import "C"

import (
	"fmt"
	"net"
	"os"

	. "tsunami"
	. "tsunami/server"
)

var (
	buildTime = C.GoString(C.build_time())
	id        uint32
)

func handler(session *Session, param *Parameter, conn *net.TCPConn) {
	defer conn.Close()

	err := session.Negotiate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Protocol revision number mismatch", err)
		return
	}

	err = session.Authenticate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Client authentication failure", err)
		return
	}

	param.VerboseArg("Client authenticated. Negotiated parameters are:")

	for {
		err = session.OpenTransfer()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid file request", err)
			continue
		}
		err = session.OpenPort()
		if err != nil {
			fmt.Fprintln(os.Stderr, "UDP socket creation failed", err)
			continue
		}
		session.Transfer()
	}
}

func newConnection(param *Parameter, conn *net.TCPConn) {
	fmt.Fprintln(os.Stderr, "New client connecting from", conn.RemoteAddr())
	session := NewSession(id, conn, param)
	id++
	go handler(session, param, conn)
}

func main() {
	param := ProcessOptions()

	ln, err := Listen(param)
	if err != nil {
		return
	}

	fmt.Fprintf(os.Stderr,
		"Tsunami Server for protocol rev %X\nRevision: %s\nCompiled: %s\n Waiting for clients to connect.\n",
		PROTOCOL_REVISION, TSUNAMI_CVS_BUILDNR, buildTime)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not accept client connection", err)
			continue
		}
		c, _ := conn.(*net.TCPConn)
		newConnection(param, c)
	}

}
