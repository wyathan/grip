package griptests

import (
	"encoding/base64"
	"errors"
	"log"
	"reflect"
	"sync"

	"github.com/wyathan/grip"
	"github.com/wyathan/grip/gripdata"
)

func testTestSocketImplements() {
	var t TestSocket
	var s grip.Socket
	s = &t
	s.Close()
}

func testTestConnectionImplements() {
	var c TestConnection
	var s grip.Connection
	s = &c
	s.Close()
}

func InitTestNetwork() *TestNetwork {
	var n TestNetwork
	n.testsockets = make(map[string]*TestSocket)
	return &n
}

type TestNetwork struct {
	sync.Mutex
	testsockets map[string]*TestSocket
}

func (n *TestNetwork) getAllSockets() []*TestSocket {
	n.Lock()
	defer n.Unlock()
	var t []*TestSocket
	for _, v := range n.testsockets {
		t = append(t, v)
	}
	return t
}

func (n *TestNetwork) CloseAll() {
	sl := n.getAllSockets()
	for _, v := range sl {
		v.Close()
	}
}

func (n *TestNetwork) Close(id []byte) {
	n.Lock()
	defer n.Unlock()
	delete(n.testsockets, base64.StdEncoding.EncodeToString(id))
}

func (n *TestNetwork) Open(id []byte, i int) *TestSocket {
	n.Lock()
	defer n.Unlock()
	var t TestSocket
	t.ID = id
	t.SC = make(chan grip.Connection, 2)
	t.Network = n
	t.Index = i
	k := base64.StdEncoding.EncodeToString(id)
	n.testsockets[k] = &t
	return &t
}

func (n *TestNetwork) GetSocket(id []byte) *TestSocket {
	n.Lock()
	defer n.Unlock()
	k := base64.StdEncoding.EncodeToString(id)
	return n.testsockets[k]
}

type TestSocket struct {
	Index   int
	ID      []byte
	SC      chan grip.Connection
	Network *TestNetwork
}

func (s *TestSocket) Close() {
	s.Network.Close(s.ID)
	close(s.SC)
}

func (s *TestSocket) Accept() (grip.Connection, error) {
	r, ok := <-s.SC
	if !ok || r == nil {
		return nil, errors.New("Socket closed")
	}
	return r, nil
}

func (s *TestSocket) ConnectTo(n *gripdata.Node) (grip.Connection, error) {
	ts := s.Network.GetSocket(n.ID)
	if ts == nil {
		return nil, errors.New("Could not connect")
	}

	var t, tr TestConnection
	t.ReadC = make(chan interface{}, 2)
	t.WriteC = make(chan interface{}, 2)
	t.ID = n.ID
	t.LclIndex = s.Index
	t.RmtIndex = ts.Index

	tr.ReadC = t.WriteC
	tr.WriteC = t.ReadC
	tr.ID = s.ID
	tr.LclIndex = ts.Index
	tr.RmtIndex = s.Index

	ts.SC <- &tr

	log.Printf("CONNECTTO: NODE: %d is connecting to node %d\n", s.Index, ts.Index)

	return &t, nil
}

type TestConnection struct {
	LclIndex int
	RmtIndex int
	ID       []byte
	WriteC   chan interface{}
	ReadC    chan interface{}
	Closed   bool
}

func (c *TestConnection) Read() (interface{}, error) {
	r, ok := <-c.ReadC
	if !ok || r == nil {
		return nil, errors.New("Connection closed")
	}
	log.Printf("READ FROM: %d TO: %d %s\n", c.RmtIndex, c.LclIndex, reflect.TypeOf(r).String())
	return r, nil
}

func (c *TestConnection) Send(d interface{}) error {
	if c.Closed {
		return errors.New("Connection closed")
	}
	c.WriteC <- d
	log.Printf("SEND FROM: %d TO: %d %s\n", c.LclIndex, c.RmtIndex, reflect.TypeOf(d).String())
	return nil
}

func (c *TestConnection) Close() {
	c.Closed = true
	close(c.WriteC)
	log.Printf("CLOSE: FROM %d TO: %d\n", c.RmtIndex, c.LclIndex)
}

func (c *TestConnection) GetNodeID() []byte {
	return c.ID
}
