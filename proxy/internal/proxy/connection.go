package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// Pre-defined error
var (
	ErrConnectionClosed  = errors.New("connection closed")
	ErrConnectionTimeout = errors.New("connection timeout")
	ErrMissingHandler    = errors.New("handler not registered")
	ErrRead              = errors.New("cannot read data from tcp connection")
	ErrWrite             = errors.New("cannot write data into tcp connection")
)

const (
	prefixSize     = 4      // 4 bytes
	maxQueueLength = 10_000 // max 10,000 connection requests in queue
)

func NewConnection(host string, port string, conn net.Conn, id string) *Connection {
	return &Connection{
		host: host,
		port: port,
		conn: conn,
		id:   id,
	}
}

type Connection struct {
	id   string
	conn net.Conn
	host string
	port string
	pool *TcpConnPool
}

type connRequest struct {
	connChan chan *Connection
	errChan  chan error
}

type TcpConnPool struct {
	host         string
	port         int
	mu           sync.Mutex             // mutex to prevent race conditions
	idleConns    map[string]*Connection // holds the idle connections
	numOpen      int                    // counter that tracks open connections
	maxOpenCount int
	maxIdleCount int
	requestChan  chan *connRequest
}

// TcpConfig is a set of configuration for a TCP connection pool
type TcpConfig struct {
	Host         string
	Port         int
	MaxIdleConns int
	MaxOpenConn  int
}

func (r *Connection) Handle() error {
	address := fmt.Sprintf("%s%s", r.host, r.port)
	mysql, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("Failed to connection to MySQL: [%d] %s", r.id, err.Error())
		return err
	}

	go func() {
		copied, err := io.Copy(mysql, r.conn)
		if err != nil {
			log.Printf("Connection error: [%d] %s", r.id, err.Error())
		}

		log.Printf("Connection closed. Bytes copied: [%d] %d", r.id, copied)
	}()

	copied, err := io.Copy(r.conn, mysql)
	if err != nil {
		log.Printf("Connection error: [%d] %s", r.id, err.Error())
		return err
	}

	log.Printf("Connection closed. Bytes copied: [%d] %d", r.id, copied)
	return nil
}