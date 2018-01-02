package adsc

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type TCPClient struct {
	tlsConfig *tls.Config
	parser    *Parser

	listenersmu   sync.Mutex
	listeners     map[uint64]*listener
	curListenerID uint64

	conn net.Conn
}

type Option func(c *TCPClient)

type listener struct {
	id uint64
	c  chan interface{}
}

func NewTCPClient(addr string, opts ...Option) (*TCPClient, error) {
	tcpc := &TCPClient{
		listeners: make(map[uint64]*listener),
	}

	for _, opt := range opts {
		opt(tcpc)
	}

	if tcpc.parser == nil {
		tcpc.parser = NewParser()
	}

	var conn net.Conn
	var err error

	if tcpc.tlsConfig != nil {
		conn, err = tls.Dial("tcp", addr, tcpc.tlsConfig)
		if err != nil {
			return nil, err
		}
	} else {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
	}

	// ensure the banner is present
	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		_ = conn.Close()

		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			return nil, errors.New("timed out reading connection banner. client/server TLS mismatch?")
		}
		return nil, err
	}
	if !strings.HasPrefix(line, "!SER2SOCK") {
		_ = conn.Close()
		return nil, errors.New("did not see ser2sock connection message")
	}

	conn.SetReadDeadline(time.Time{})
	tcpc.conn = conn
	go tcpc.readMessages()
	return tcpc, nil
}

func WithTLSConfig(conf *tls.Config) Option {
	return func(client *TCPClient) {
		client.tlsConfig = conf
	}
}

func WithParser(p *Parser) Option {
	return func(client *TCPClient) {
		client.parser = p
	}
}

func (c *TCPClient) Messages() (chan interface{}, func()) {
	c.listenersmu.Lock()
	defer c.listenersmu.Unlock()
	c.curListenerID++
	listener := &listener{
		id: c.curListenerID,
		c:  make(chan interface{}),
	}
	c.listeners[c.curListenerID] = listener
	return listener.c, func() { c.stopListener(c.curListenerID) }
}

func (c *TCPClient) readMessages() {
	rdr := bufio.NewReader(c.conn)
	for {
		line, err := rdr.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")
		if err != nil {
			fmt.Printf("%v %s\n", err, line)
			continue
		}

		msg, err := c.parser.Parse(line)
		if err != nil && err != ErrUnhandled {
			fmt.Printf("%v %s\n", err, line)
			continue
		}

		if msg != nil {
			for _, l := range c.listeners {
				l.c <- msg
			}
		}
	}
}

func (c *TCPClient) stopListener(id uint64) {
	c.listenersmu.Lock()
	defer c.listenersmu.Unlock()
	delete(c.listeners, id)
}
