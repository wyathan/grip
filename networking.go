package grip

import (
	"encoding/base64"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/wyathan/grip/gripdata"
)

//MAXSEND The maximum number of SendData to query from
//the database and send on the connection at one time
const MAXSEND int = 10

//SLEEPONNOSEND The number of seconds to sleep between
//querying the database for new data to send
const SLEEPONNOSEND time.Duration = 1 * time.Second

//BUFFERSIZE is the send channel size
const BUFFERSIZE int = 2

//MAXCONNECTIONS is the maximum number of connections we allow
const MAXCONNECTIONS int = 1000

//MAXCONNECTIONATTEMPTS number of connections to attempt at once
const MAXCONNECTIONATTEMPTS int = 20

//SLEEPONNOCONNECTIONS if we run out of connections how long do
//we wait until we test it again
const SLEEPONNOCONNECTIONS time.Duration = 5 * time.Second

//Connection handles sending and reading data
//on a connection
type Connection interface {
	Read() (interface{}, error)
	Send(d interface{}) error
	GetNodeID() []byte //The validated id of the node connected to
	Close()            //Must cause blocking read to immediately exit
}

//Socket handles making connections and accepting them
type Socket interface {
	ConnectTo(n *gripdata.Node) (Connection, error)
	Accept() (Connection, error)
	Close()
}

//SocketController handles a socket
type SocketController struct {
	sync.Mutex
	Connections map[string]*ConnectionController
	S           Socket
	Done        bool
	DB          NodeNetAccountdb
}

func NewSocketController(sock Socket, db NodeNetAccountdb) *SocketController {
	var s SocketController
	s.S = sock
	s.DB = db
	s.Connections = make(map[string]*ConnectionController)
	return &s
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
		c.SocketCtrl = nil
		c.Close()
		return false
	}
	c.SocketCtrl = s
	s.Connections[ck] = c
	return true
}

func (s *SocketController) removeConnection(id []byte) {
	s.Lock()
	defer s.Unlock()
	ck := base64.StdEncoding.EncodeToString(id)
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
	cl := s.DB.GetAllConnected()
	for _, c := range cl {
		c.Connected = false
		err := s.DB.StoreNodeEphemera(&c)
		if err != nil {
			log.Fatal("Failed to update database")
		}
	}
	go s.listenRoutine()
	go s.connectRoutine()
}

func (s *SocketController) listenRoutine() {
	con, err := s.S.Accept()
	for !s.Done && err == nil {
		if con.GetNodeID() == nil {
			log.Printf("This is bad.  Connections should only return with valid ID")
		} else {
			var ctrl ConnectionController
			ctrl.C = con
			ctrl.DB = s.DB
			ctrl.Incoming = true
			ctrl.SendChan = make(chan interface{}, BUFFERSIZE)
			s.addConnection(&ctrl)
			go ctrl.ConnectionReadRoutine()
			go ctrl.ConnectionWriteRoutine()
		}
		con, err = s.S.Accept()
	}
}

func (s *SocketController) connectRoutine() {
	log.Printf("Starting connectRoutine")
	for !s.Done {
		//Check max connections
		nm := s.numberConnections()
		if nm <= MAXCONNECTIONS {
			cl := s.DB.GetConnectableNodesWithSendData(MAXCONNECTIONATTEMPTS, uint64(time.Now().UnixNano()))
			for _, c := range cl {
				n := s.DB.GetNode(c.ID)
				con, err := s.S.ConnectTo(n)
				if con != nil && err == nil {
					var ctrl ConnectionController
					ctrl.C = con
					ctrl.DB = s.DB
					ctrl.Incoming = false
					ctrl.SendChan = make(chan interface{}, BUFFERSIZE)
					s.addConnection(&ctrl)
					go ctrl.ConnectionReadRoutine()
					go ctrl.ConnectionWriteRoutine()
				} else {
					nt := uint64(time.Now().UnixNano())
					c.LastConnAttempt = nt
					c.NextAttempt = nt + ((nt - c.LastConnection) / 2)
					s.DB.StoreNodeEphemera(&c)
				}
			}
		}
	}
}

//ConnectionController is a handle for connections
type ConnectionController struct {
	sync.Mutex
	C          Connection
	SendChan   chan interface{}
	Done       bool
	Incoming   bool
	DB         NodeNetAccountdb
	SocketCtrl *SocketController
}

//CheckDig check if a node has a digest you want to send it
type CheckDig struct {
	Dig []byte
}

//RespDig lets other node know if we have this object yet or not
type RespDig struct {
	HaveIt bool
	Dig    []byte
}

//SendDig actually send this data
type SendDig struct {
	Dig []byte
}

//Close a connection
func (ctrl *ConnectionController) Close() {
	ctrl.Lock()
	defer ctrl.Unlock()
	if !ctrl.Done {
		ctrl.Done = true
		close(ctrl.SendChan)
		if ctrl.SocketCtrl != nil {
			ctrl.SocketCtrl.removeConnection(ctrl.C.GetNodeID())
		}
	}
}

func (ctrl *ConnectionController) sendFromDatabase() (int, error) {
	c := 0
	sl := ctrl.DB.GetSendData(ctrl.C.GetNodeID(), MAXSEND)
	err := ctrl.sendSendDataList(sl)
	if err != nil {
		return 0, err
	}
	c += len(sl)
	return c, nil
}

func (ctrl *ConnectionController) sendSendDataList(sl []gripdata.SendData) error {
	for _, v := range sl {
		var cd CheckDig
		cd.Dig = v.Dig
		ctrl.C.Send(cd)
	}
	return nil
}

func (ctrl *ConnectionController) sendSendData(d []byte) error {
	sd := ctrl.DB.GetDigestData(d)
	if sd != nil {
		err := ctrl.C.Send(sd)
		if err != nil {
			return err
		}
		err = ctrl.deleteSendData(d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *ConnectionController) deleteSendData(d []byte) error {
	return ctrl.DB.DeleteSendData(d, ctrl.C.GetNodeID())
}

//ConnectionWriteRoutine go routine to write to connections
func (ctrl *ConnectionController) ConnectionWriteRoutine() {
	defer ctrl.C.Close() //Only close connection here

	if !ctrl.Done {
		ctrl.setConnected(ctrl.Incoming)
		myn, _ := ctrl.DB.GetPrivateNodeData()
		ctrl.C.Send(myn)
		c, err := ctrl.sendFromDatabase()
		for err == nil && !ctrl.Done {
			timeout := make(chan bool, 1)
			go func(ct int) {
				if ct == 0 {
					time.Sleep(SLEEPONNOSEND)
				}
				timeout <- true
			}(c)
			select {
			case r, ok := <-ctrl.SendChan:
				if !ok || ctrl.Done {
					//Do not call ctrl.Close() because
					//its only purpose is to close ctlr.S
					//so that this routine exits.  We're
					//already exiting this routine and
					//the connection will be closed.
					ctrl.Done = true
				} else {
					switch v := r.(type) {
					case SendDig:
						err = ctrl.sendSendData(v.Dig)
					case bool:
						c, err = ctrl.sendFromDatabase()
					default:
						err = ctrl.C.Send(v)
					}
				}
			case <-timeout:
				c, err = ctrl.sendFromDatabase()
			}
		}
		if err != nil {
			log.Printf("Connection error: %s\n", err)
		}
		ctrl.setConnectionClosed()
	}
}

//ConnectionReadRoutine go routine to read from connections
func (ctrl *ConnectionController) ConnectionReadRoutine() {
	defer ctrl.Close()
	d, err := ctrl.C.Read()
	for err == nil {
		if d != nil {
			log.Printf("Incoming type: %s\n", reflect.TypeOf(d).String())
			switch v := d.(type) {
			default:
				log.Printf("Unknown type: %s\n", reflect.TypeOf(v).String())
			case CheckDig:
				t := ctrl.DB.GetDigestData(v.Dig)
				var rsp RespDig
				rsp.Dig = v.Dig
				rsp.HaveIt = (t != nil)
				ctrl.SendChan <- rsp
			case RespDig:
				if v.HaveIt {
					ctrl.deleteSendData(v.Dig)
				} else {
					var sd SendDig
					sd.Dig = v.Dig
					ctrl.SendChan <- sd
				}
			case *gripdata.Node:
				log.Println("Node data received")
				err = IncomingNode(v, ctrl.DB)
			case *gripdata.AssociateNodeAccountKey:
				log.Println("AssociateNodeAccountKey data received")
				err = IncomingNodeAccountKey(v, ctrl.DB)
			case *gripdata.UseShareNodeKey:
				log.Println("UseShareNodeKey data received")
				err = IncomingUseShareNodeKey(v, ctrl.DB)
			case *gripdata.ShareNodeInfo:
				log.Println("ShareNodeInfo data received")
				err = IncomingShareNode(v, ctrl.DB)
			}
		}
		d, err = ctrl.C.Read()
	}
	if err != nil {
		log.Printf("Connection error: %s\n", err)
	}
}

func (ctrl *ConnectionController) setConnected(incomming bool) error {
	nid := ctrl.DB.GetNodeEphemera(ctrl.C.GetNodeID())
	if nid == nil {
		var n gripdata.NodeEphemera
		n.ID = ctrl.C.GetNodeID()
		nid = &n
	}
	nid.Connected = true
	if incomming {
		nid.LastConnReceived = uint64(time.Now().UnixNano())
	} else {
		nid.LastConnection = uint64(time.Now().UnixNano())
	}
	return ctrl.DB.StoreNodeEphemera(nid)
}

func (ctrl *ConnectionController) setConnectionClosed() {
	nid := ctrl.DB.GetNodeEphemera(ctrl.C.GetNodeID())
	if nid != nil {
		nid.Connected = false
		ctrl.DB.StoreNodeEphemera(nid)
	}
}
