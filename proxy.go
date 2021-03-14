package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func forward(src net.Conn, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}

func handleProxyConnection(conn net.Conn, dstHostname string, dstPort int) {
	remote, err := net.Dial("tcp", fmt.Sprintf("%s:%d", dstHostname, dstPort))
	if err != nil {
		log.Println("Proxy dial error: " + err.Error())
		return
	}

	go forward(conn, remote)
	go forward(remote, conn)
}

func proxy(srcHostname string, srcPort int, dstHostname string, dstPort int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", srcHostname, srcPort))
	if err != nil {
		log.Println("Listen error: " + err.Error())
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error on accepting proxy connection:", err)
			continue
		}

		log.Printf("Proxying connection %s:%d -> %s:%d\n", srcHostname, srcPort, dstHostname, dstPort)
		go handleProxyConnection(conn, dstHostname, dstPort)
	}
}
