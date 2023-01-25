package main

import (
	"flag"
	"fmt"
	"log"
	"net"

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

	pool, poolError := proxy.CreateTcpConnPool(&proxy.TcpConfig{
		Host:         *flag.String("tcp-host", "127.0.0.1", "host of tcp server"),
		Port:         *flag.Int("tcp-port", 3306, "port of tcp server"),
		MaxIdleConns: *flag.Int("max-idle", 1, "max number of idle tcp conns"),
		MaxOpenConn:  *flag.Int("max-open", 3, "max number of open tcp conns"),
	})

	if poolError != nil {
		log.Printf("Failed to accept new connection: [%d] %s", r.connectionId, poolError.Error())

	}

	for {
		conn, err := ln.Accept()

		// r.connectionId = fmt.Sprintf("%v", time.Now().UnixNano())
		log.Printf("Connection accepted: [%s] %s", r.connectionId, conn.RemoteAddr())

		if err != nil {
			log.Printf("Failed to accept new connection: [%s] %s", r.connectionId, err.Error())
			continue
		}

		go r.handle(conn, pool)
	}
}

func (r *Proxy) handle(conn net.Conn, pool *proxy.TcpConnPool) {
	connection := proxy.NewConnection(conn, pool)

	err := connection.Handle()
	if err != nil {
		log.Printf("Error handling proxy connection: %s", err.Error())
	}
}
