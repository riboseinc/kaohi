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
 	"io"
 	"net"
 	"sync"
 	"fmt"
	"encoding/binary"
	"errors"
	"sync/atomic"
	"time"
)

const (
	MAX_CONN_COUNT = 1024
)

// kaohi command listener struct
type KaohiCmdListener struct {
	listener *net.TCPListener
	srv      *CmdServer
}

// global variable for command listener
var kCmdListener KaohiCmdListener

type CmdConfig struct {
	PacketSendChanLimit    uint32 // the limit of packet send channel
	PacketReceiveChanLimit uint32 // the limit of packet receive channel
}

type CmdServer struct {
	config    *CmdConfig         // server configuration
	callback  ConnCallback       // message callbacks in connection
	protocol  CmdProtocol        // customize packet protocol
	exitChan  chan struct{}      // notify all goroutines to shutdown
	waitGroup *sync.WaitGroup    // wait for all goroutines
}

type CmdPacket struct {
	buff []byte
}

func (this *CmdPacket) Serialize() []byte {
	return this.buff
}

func (this *CmdPacket) GetLength() uint32 {
	return binary.BigEndian.Uint32(this.buff[0:4])
}

func (this *CmdPacket) GetBody() []byte {
	return this.buff[4:]
}

func NewEchoPacket(buff []byte, hasLengthField bool) (p *CmdPacket) {
	if hasLengthField {
		p.buff = buff

	} else {
		p.buff = make([]byte, 4+len(buff))
		binary.BigEndian.PutUint32(p.buff[0:4], uint32(len(buff)))
		copy(p.buff[4:], buff)
	}

	return p
}

type CmdProtocol struct {
}

func (this *CmdProtocol) ReadPacket(conn *net.TCPConn) (*CmdPacket, error) {
	var (
		lengthBytes []byte = make([]byte, 4)
		length      uint32
	)

	// read length
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		return nil, err
	}
	if length = binary.BigEndian.Uint32(lengthBytes); length > 1024 {
		return nil, errors.New("the size of packet is larger than the limit")
	}

	buff := make([]byte, 4+length)
	copy(buff[0:4], lengthBytes)

	// read body ( buff = lengthBytes + body )
	if _, err := io.ReadFull(conn, buff[4:]); err != nil {
		return nil, err
	}

	return NewEchoPacket(buff, true), nil
}

// Conn exposes a set of callbacks for the various events that occur on a connection
type Conn struct {
	srv               *CmdServer
	conn              *net.TCPConn      // the raw connection
	extraData         interface{}       // to save extra data
	closeOnce         sync.Once         // close the conn, once, per instance
	closeFlag         int32             // close flag
	closeChan         chan struct{}     // close chanel
	packetSendChan    chan *CmdPacket   // packet send chanel
	packetReceiveChan chan *CmdPacket   // packeet receive chanel
}

// ConnCallback is an interface of methods that are used as callbacks on a connection
type ConnCallback interface {
	// OnConnect is called when the connection was accepted,
	// If the return value of false is closed
	OnConnect(*Conn) bool

	// OnMessage is called when the connection receives a packet,
	// If the return value of false is closed
	OnMessage(*Conn, *CmdPacket) bool

	// OnClose is called when the connection closed
	OnClose(*Conn)
}

// newConn returns a wrapper of raw conn
func newConn(conn *net.TCPConn, srv *CmdServer) *Conn {
	DEBUG_INFO("Accepted new connection")

	return &Conn{
		srv:               srv,
		conn:              conn,
		closeChan:         make(chan struct{}),
		packetSendChan:    make(chan *CmdPacket, srv.config.PacketSendChanLimit),
		packetReceiveChan: make(chan *CmdPacket, srv.config.PacketReceiveChanLimit),
	}
}

// GetExtraData gets the extra data from the Conn
func (c *Conn) GetExtraData() interface{} {
	return c.extraData
}

// PutExtraData puts the extra data with the Conn
func (c *Conn) PutExtraData(data interface{}) {
	c.extraData = data
}

// GetRawConn returns the raw net.TCPConn from the Conn
func (c *Conn) GetRawConn() *net.TCPConn {
	return c.conn
}

// Close closes the connection
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		close(c.packetSendChan)
		close(c.packetReceiveChan)
		c.conn.Close()
		c.srv.callback.OnClose(c)
	})
}

// IsClosed indicates whether or not the connection is closed
func (c *Conn) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}

// AsyncWritePacket async writes a packet, this method will never block
func (c *Conn) AsyncWritePacket(p *CmdPacket, timeout time.Duration) (err error) {
	if c.IsClosed() {
		return ErrConnClosing
	}

	defer func() {
		if e := recover(); e != nil {
			err = ErrConnClosing
		}
	}()

	if timeout == 0 {
		select {
		case c.packetSendChan <- p:
			return nil

		default:
			return ErrWriteBlocking
		}

	} else {
		select {
		case c.packetSendChan <- p:
			return nil

		case <-c.closeChan:
			return ErrConnClosing

		case <-time.After(timeout):
			return ErrWriteBlocking
		}
	}
}

// Do it
func (c *Conn) Do() {
	if !c.srv.callback.OnConnect(c) {
		return
	}

	asyncDo(c.handleLoop, c.srv.waitGroup)
	asyncDo(c.readLoop, c.srv.waitGroup)
	asyncDo(c.writeLoop, c.srv.waitGroup)
}

func (c *Conn) readLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		default:
		}

		p, err := c.srv.protocol.ReadPacket(c.conn)
		if err != nil {
			return
		}

		c.packetReceiveChan <- p
	}
}

func (c *Conn) writeLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetSendChan:
			if c.IsClosed() {
				return
			}
			if _, err := c.conn.Write(p.Serialize()); err != nil {
				return
			}
		}
	}
}

func (c *Conn) handleLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetReceiveChan:
			if c.IsClosed() {
				return
			}
			if !c.srv.callback.OnMessage(c, p) {
				return
			}
		}
	}
}

func asyncDo(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()
}

// NewServer creates a server
func NewServer(config *CmdConfig, callback ConnCallback, protocol CmdProtocol) *CmdServer {
	return &CmdServer{
		config:    config,
		callback:  callback,
		protocol:  protocol,
		exitChan:  make(chan struct{}),
		waitGroup: &sync.WaitGroup{},
	}
}

// Start starts service
func (s *CmdServer) Start(listener *net.TCPListener, acceptTimeout time.Duration) {
	DEBUG_INFO("Starting command listening service")

	s.waitGroup.Add(1)
	defer func() {
		listener.Close()
		s.waitGroup.Done()
	}()

	for {
		select {
		case <-s.exitChan:
			return

		default:
		}

		listener.SetDeadline(time.Now().Add(acceptTimeout))
		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}

		s.waitGroup.Add(1)
		go func() {
			newConn(conn, s).Do()
			s.waitGroup.Done()
		}()
	}
}

// Stop stops service
func (s *CmdServer) Stop() {
	close(s.exitChan)
	s.waitGroup.Wait()
}

type Callback struct{}

func (this *Callback) OnConnect(c *Conn) bool {
	addr := c.GetRawConn().RemoteAddr()
	c.PutExtraData(addr)
	return true
}

func (this *Callback) OnMessage(c *Conn, p *CmdPacket) bool {
	fmt.Printf("OnMessage:[%v] [%v]\n", p.GetLength(), string(p.GetBody()))
	c.AsyncWritePacket(NewEchoPacket(p.Serialize(), true), time.Second)
	return true
}

func (this *Callback) OnClose(c *Conn) {2
	fmt.Println("OnClose:", c.GetExtraData())
}

// init kaohi command listener
func InitCmdListener(ctx *kContext) error {
	var err error

	DEBUG_INFO("Initializing command listener")

	// create listener
	listenAddr, err := net.ResolveTCPAddr("tcp4", ctx.config.listen_addr)
	if err != nil {
		return ErrResolveAddr
	}

	kCmdListener.listener, err = net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return ErrListenFaield
	}

	DEBUG_INFO("Listening on %s", ctx.config.listen_addr)

	// creates a server instance
	config := &CmdConfig{
		PacketReceiveChanLimit: 20,
	}
	kCmdListener.srv = NewServer(config, &Callback{}, CmdProtocol{})

	// starts service
	go kCmdListener.srv.Start(kCmdListener.listener, time.Second)

	return nil
}

// finalize kaohi command listener
func FinalizeCmdListener() {
	DEBUG_INFO("Finalzing command listener")

	kCmdListener.srv.Stop()
}
