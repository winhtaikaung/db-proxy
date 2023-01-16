package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/db-proxy/proxy/internal/proxy"
)

func NewProxy(host, port string) *Proxy {
	return &Proxy{
		host: host,
		port: port,
	}
}

type Proxy struct {
	host         string
	port         string
	connectionId string
}

// func getResolvedAddresses(host string) *net.TCPAddr {
// 	addr, err := net.ResolveTCPAddr("tcp", host)
// 	if err != nil {
// 		log.Fatalln("ResolveTCPAddr of host:", err)
// 	}
// 	return addr
// }

// // Listener of a net.TCPAddr.
// func getListener(addr *net.TCPAddr) *net.TCPListener {
// 	listener, err := net.ListenTCP("tcp", addr)
// 	if err != nil {
// 		log.Fatalf("ListenTCP of %s error:%v", addr, err)
// 	}
// 	return listener
// }

func (r *Proxy) Start(port string) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		r.connectionId = fmt.Sprintf("%v", time.Now().UnixNano())
		log.Printf("Connection accepted: [%d] %s", r.connectionId, conn.RemoteAddr())
		if err != nil {
			log.Printf("Failed to accept new connection: [%d] %s", r.connectionId, err.Error())
			continue
		}

		go r.handle(conn, r.connectionId)
	}
}

func (r *Proxy) handle(conn net.Conn, connectionId string) {
	connection := proxy.NewConnection(r.host, r.port, conn, connectionId)
	err := connection.Handle()
	if err != nil {
		log.Printf("Error handling proxy connection: %s", err.Error())
	}
}
