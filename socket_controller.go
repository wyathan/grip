package grip

import (
	"encoding/base64"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/orcaman/concurrent-map"
)

//SocketController handles a socket
type SocketController struct {
	sync.Mutex
	Connections map[string]*ConnectionController
	S           Socket
	Done        bool
	DB          NodeNetAccountContextdb
	LastLoop    uint64
}

//NewSocketController builds a new SocketController to handle
//incoming connections
func NewSocketController(sock Socket, db NodeNetAccountContextdb) *SocketController {
	var s SocketController
	s.S = sock
	s.DB = db
	s.Connections = make(map[string]*ConnectionController)
	return &s
}
func (s *SocketController) checkEphemera(mstr string) {
	conlst := s.DB.GetAllConnected()
	for _, cn := range conlst {
		st := base64.StdEncoding.EncodeToString(cn.ID)
		con := s.Connections[st]
		if con == nil {
			log.Printf("ERROR: Node Ephemera says connected, but not! me %s, from %s", mstr, st)
		}
	}
}
func (s *SocketController) checkConnections() {
	s.Lock()
	defer s.Unlock()
	mid, _ := s.DB.GetPrivateNodeData()
	mstr := base64.StdEncoding.EncodeToString(mid.ID)
	for k, con := range s.Connections {
		if con.Done {
			log.Printf("ERROR: Connection done, but still connected me: %s, from: %s", mstr, k)
		}
	}
	s.checkEphemera(mstr)
}

func (s *SocketController) addConnection(c *ConnectionController) bool {
	s.Lock()
	defer s.Unlock()
	if len(s.Connections) >= MAXCONNECTIONS {
		c.SocketCtrl = nil
		c.Close()
		return false
	}
	ck := base64.StdEncoding.EncodeToString(c.C.GetNodeID())
	cc := s.Connections[ck]
	if cc != nil {
		//NOTE00 This is important otherwise the existing connection
		//may be removed from the Connections map when this
		//one is closed
		c.SocketCtrl = nil
		c.Close()
		return false
	}
	c.SocketCtrl = s
	log.Printf("ADDCON!")
	s.Connections[ck] = c
	return true
}
func (s *SocketController) removeConnection(id []byte) {
	s.Lock()
	defer s.Unlock()
	ck := base64.StdEncoding.EncodeToString(id)
	log.Printf("DELETE")
	delete(s.Connections, ck)
}
func (s *SocketController) checkConnection(id []byte) bool {
	s.Lock()
	defer s.Unlock()
	ck := base64.StdEncoding.EncodeToString(id)
	return nil != s.Connections[ck]
}
func (s *SocketController) numberConnections() int {
	s.Lock()
	defer s.Unlock()
	return len(s.Connections)
}
func (s *SocketController) listConnections() []*ConnectionController {
	s.Lock()
	defer s.Unlock()
	var r []*ConnectionController
	for _, v := range s.Connections {
		r = append(r, v)
	}
	return r
}

//Close everything!  It's over man
func (s *SocketController) Close() {
	cg := s.closeGate()
	if cg {
		s.closeAllConnections()
		s.S.Close()
	}
}
func (s *SocketController) closeGate() bool {
	s.Lock()
	defer s.Unlock()
	if !s.Done {
		s.Done = true
		return true
	}
	return false
}
func (s *SocketController) closeAllConnections() {
	cl := s.listConnections()
	for _, c := range cl {
		c.Close()
	}
}

//Start begin the goroutines for the SocketController
func (s *SocketController) Start() {
	//We're just starting.  We can't be
	//connected to any node
	s.DB.ClearAllConnected()
	go s.listenRoutine()
	go s.connectRoutine()
}

func (s *SocketController) buildConnectionController(con Connection, incomming bool) {
	var ctrl ConnectionController
	ctrl.C = con
	ctrl.DB = s.DB
	ctrl.Incoming = incomming
	ctrl.SendChan = make(chan interface{}, BUFFERSIZE)
	ctrl.Pending = cmap.New()
	ctrl.ConID = rand.Uint64()
	s.addConnection(&ctrl)
	go ctrl.ConnectionReadRoutine()
	go ctrl.ConnectionWriteRoutine()
}

func (s *SocketController) listenRoutine() {
	con, err := s.S.Accept()
	for !s.Done && err == nil {
		if con.GetNodeID() == nil {
			log.Printf("This is bad.  Connections should only return with valid ID")
		} else {
			if s.DB.CanNodeEphemeraGoPending(con.GetNodeID()) {
				s.buildConnectionController(con, true)
			} else {
				con.Close()
				s.DB.SetNodeEphemeraClosed(con.GetNodeID())
			}
		}
		con, err = s.S.Accept()
		s.LastLoop = uint64(time.Now().UnixNano())
	}
}
