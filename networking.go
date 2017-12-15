package grip

import (
	"encoding/base64"
	"errors"
	"log"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/orcaman/concurrent-map"

	"github.com/wyathan/grip/gripdata"
)

//MAXSEND The maximum number of SendData to query from
//the database and send on the connection at one time
const MAXSEND int = 10

//SLEEPONNOSEND The number of seconds to sleep between
//querying the database for new data to send
const SLEEPONNOSEND time.Duration = 1 * time.Second

//BUFFERSIZE is the send channel size
const BUFFERSIZE int = 20

//MAXCONNECTIONS is the maximum number of connections we allow
const MAXCONNECTIONS int = 1000

//MAXCONNECTIONATTEMPTS number of connections to attempt at once
const MAXCONNECTIONATTEMPTS int = 20

//TRYCONNECTAGAINAFTERFAIL how long to wait until we try to connect
//again after we've failed
const TRYCONNECTAGAINAFTERFAIL time.Duration = 1 * time.Second

//WAITUNTILCONNECTAGAIN give connections this amount of time to
//become established and upated in the database before we try
//to connect again.  It should take less time than this for both
//ends of a connection to close after one side closes, otherwise
//existing connections that haven't yet completely closed will
//keep new connections from being made.
const WAITUNTILCONNECTAGAIN time.Duration = TRYCONNECTAGAINAFTERFAIL

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

func (s *SocketController) checkConnections() {
	s.Lock()
	defer s.Unlock()
	mid, _ := s.DB.GetPrivateNodeData()
	for k, con := range s.Connections {
		if con.Done {
			log.Printf("ERROR: Connection done, but still connected me: %s, from: %s",
				base64.StdEncoding.EncodeToString(mid.ID), k)
		}
	}
	conlst := s.DB.GetAllConnected()
	for _, cn := range conlst {
		st := base64.StdEncoding.EncodeToString(cn.ID)
		con := s.Connections[st]
		if con == nil {
			log.Printf("ERROR: Node Ephemera says connected, but not! me %s, from %s",
				base64.StdEncoding.EncodeToString(mid.ID), st)
		}
	}
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

func (s *SocketController) listenRoutine() {
	con, err := s.S.Accept()
	for !s.Done && err == nil {
		if con.GetNodeID() == nil {
			log.Printf("This is bad.  Connections should only return with valid ID")
		} else {
			if s.DB.CanNodeEphemeraGoPending(con.GetNodeID()) {
				var ctrl ConnectionController
				ctrl.C = con
				ctrl.DB = s.DB
				ctrl.Incoming = true
				ctrl.SendChan = make(chan interface{}, BUFFERSIZE)
				ctrl.Pending = cmap.New()
				ctrl.ConID = rand.Uint64()
				s.addConnection(&ctrl)
				go ctrl.ConnectionReadRoutine()
				go ctrl.ConnectionWriteRoutine()
			} else {
				con.Close()
				s.DB.SetNodeEphemeraClosed(con.GetNodeID())
			}
		}
		con, err = s.S.Accept()
		s.LastLoop = uint64(time.Now().UnixNano())
	}
}

func (s *SocketController) connectToNodeEphemera(c *gripdata.NodeEphemera) {
	if s.DB.CanNodeEphemeraGoPending(c.ID) {
		n := s.DB.GetNode(c.ID)
		con, err := s.S.ConnectTo(n)
		if con != nil && err == nil {
			var ctrl ConnectionController
			ctrl.C = con
			ctrl.DB = s.DB
			ctrl.Incoming = false
			ctrl.SendChan = make(chan interface{}, BUFFERSIZE)
			ctrl.Pending = cmap.New()
			ctrl.ConID = rand.Uint64()
			s.addConnection(&ctrl)
			go ctrl.ConnectionReadRoutine()
			go ctrl.ConnectionWriteRoutine()
		} else {
			if con != nil {
				con.Close()
			}
			log.Print("Connection error\n")
			nt := uint64(time.Now().UnixNano())
			s.DB.SetNodeEphemeraNextConnection(c.ID, nt, nt+uint64(TRYCONNECTAGAINAFTERFAIL.Nanoseconds()))
		}
	}
}

func (s *SocketController) canAttempt(attempted *map[string]uint64, id []byte, nowtime uint64) bool {
	sid := base64.StdEncoding.EncodeToString(id)
	oldenough := nowtime - uint64(WAITUNTILCONNECTAGAIN)
	if (*attempted)[sid] < oldenough {
		(*attempted)[sid] = nowtime
		return true
	}
	return false
}

func (s *SocketController) connectRoutine() {
	attempted := make(map[string]uint64)
	for !s.Done {
		nt := uint64(time.Now().UnixNano())
		nm := s.numberConnections()
		if nm <= MAXCONNECTIONS {
			cl := s.DB.GetConnectableNodesWithSendData(MAXCONNECTIONATTEMPTS, nt)
			for _, c := range cl {
				if s.canAttempt(&attempted, c.ID, nt) {
					s.connectToNodeEphemera(&c)
				}
			}
		}
		nm = s.numberConnections()
		if nm <= MAXCONNECTIONS {
			cl := s.DB.GetConnectableNodesWithShareNodeKey(MAXCONNECTIONATTEMPTS, nt)
			for _, c := range cl {
				if s.canAttempt(&attempted, c.ID, nt) {
					s.connectToNodeEphemera(&c)
				}
			}
		}
		nm = s.numberConnections()
		if nm <= MAXCONNECTIONS {
			cl := s.DB.GetConnectableUseShareKeyNodes(MAXCONNECTIONATTEMPTS, nt)
			for _, c := range cl {
				if s.canAttempt(&attempted, c.ID, nt) {
					s.connectToNodeEphemera(&c)
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
		nt = uint64(time.Now().UnixNano())
		nm = s.numberConnections()
		if nm <= MAXCONNECTIONS {
			cl := s.DB.GetConnectableAny(MAXCONNECTIONATTEMPTS, nt)
			//Select one at random
			ln := len(cl)
			if ln > 0 {
				idx := rand.Intn(len(cl))
				if s.canAttempt(&attempted, cl[idx].ID, nt) {
					log.Printf("Attempt random connection")
					s.connectToNodeEphemera(&cl[idx])
				}
			}
		}
		s.checkConnections()
	}
}

//ConnectionController is a handle for connections
type ConnectionController struct {
	sync.Mutex
	C             Connection
	SendChan      chan interface{}
	Done          bool
	Incoming      bool
	DB            NodeNetAccountContextdb
	SocketCtrl    *SocketController
	Pending       cmap.ConcurrentMap
	ConID         uint64
	LastReadLoop  uint64
	LastWriteLoop uint64
}

func (s *SocketController) ShowLastLoops(m map[string]int) {
	s.Lock()
	defer s.Unlock()
	mid, _ := s.DB.GetPrivateNodeData()
	msid := base64.StdEncoding.EncodeToString(mid.ID)
	log.Printf("Node %d Last listen loop %d", m[msid], s.LastLoop)
	for k, v := range s.Connections {
		log.Printf("    To Node %d  last write %d, last read %d", m[k], v.LastWriteLoop, v.LastReadLoop)
	}
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
	HaveIt bool
	Dig    []byte
}

//RejectDig indicates this digest was rejected by the receving node
type RejectDig struct {
	Dig     []byte
	Message string
}

//AckDig indicates the receiving node got the data
type AckDig struct {
	Dig []byte
}

//Close a connection
func (ctrl *ConnectionController) Close() {
	//Empty SendChan in case sendToChan is blocking on
	//a full channel keeping us from getting the ctrl lock
	go func() {
		for nil != <-ctrl.SendChan {
		}
	}()
	ctrl.Lock()
	if !ctrl.Done {
		ctrl.Done = true
		close(ctrl.SendChan)
		if ctrl.SocketCtrl != nil {
			ctrl.SocketCtrl.removeConnection(ctrl.C.GetNodeID())
		}
	}
	ctrl.Unlock()
}

//This is nasty, should just find way of only closing SendChan once done sending
//for good
func (ctrl *ConnectionController) sendToChan(d interface{}) {
	ctrl.Lock()
	defer ctrl.Unlock()
	if !ctrl.Done {
		ctrl.SendChan <- d
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
		if ctrl.Done {
			return errors.New("Connection closed")
		}
		ds := base64.StdEncoding.EncodeToString(v.Dig)
		fnd := false
		val, ok := ctrl.Pending.Get(ds)
		if ok {
			fnd = val.(bool)
		}
		if !fnd {
			var cd CheckDig
			cd.Dig = v.Dig
			err := ctrl.C.Send(cd)
			if err != nil {
				return err
			}
			ctrl.Pending.Set(ds, true)
		}
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
	}
	return nil
}

func (ctrl *ConnectionController) deleteSendData(d []byte) error {
	defer ctrl.Pending.Remove(base64.StdEncoding.EncodeToString(d))
	log.Printf("Deleting senddata %s\n", base64.StdEncoding.EncodeToString(d))
	return ctrl.DB.DeleteSendData(d, ctrl.C.GetNodeID())
}

func (ctrl *ConnectionController) sendDataRejected(d []byte, msg string) error {
	var r RejectDig
	r.Dig = d
	r.Message = msg
	return ctrl.C.Send(r)
}

func (ctrl *ConnectionController) storeDataRejected(d []byte) error {
	r := gripdata.RejectedSendData{}
	r.Dig = d
	r.TargetID = ctrl.C.GetNodeID()
	r.Timestamp = uint64(time.Now().UnixNano())
	err := ctrl.DB.StoreRejectedSendData(&r)
	if err != nil {
		return err
	}
	return nil
}

//ConnectionWriteRoutine go routine to write to connections
func (ctrl *ConnectionController) ConnectionWriteRoutine() {
	defer func() {
		ctrl.setConnectionClosed()
		ctrl.C.Close() //Only close connection here
	}()
	if !ctrl.Done {
		ctrl.setConnected(ctrl.Incoming)
		myn, _ := ctrl.DB.GetPrivateNodeData()
		err := ctrl.C.Send(myn)
		if err != nil {
			log.Print("Failed to send my node data for connection init")
			ctrl.Close()
		}
		c, err := ctrl.sendFromDatabase()
		if err != nil {
			ctrl.Close()
		}
		for !ctrl.Done {
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
						if v.HaveIt {
							ctrl.deleteSendData(v.Dig)
						} else {
							err = ctrl.sendSendData(v.Dig)
						}
					case bool:
						c, err = ctrl.sendFromDatabase()
					default:
						err = ctrl.C.Send(v)
					}
				}
			case <-timeout:
				c, err = ctrl.sendFromDatabase()
			}
			if err != nil {
				ctrl.Close()
			}
			ctrl.LastWriteLoop = uint64(time.Now().UnixNano())
		}
	}
}

func (ctrl *ConnectionController) processSendError(fname string, dig []byte, err error) {
	if err == nil {
		log.Printf("%s data received", fname)
		ctrl.sendToChan(AckDig{Dig: dig})
	} else {
		log.Printf("%s data rejected: %s", fname, err)
		ctrl.sendDataRejected(dig, err.Error())
	}
}

//ConnectionReadRoutine go routine to read from connections
func (ctrl *ConnectionController) ConnectionReadRoutine() {
	defer ctrl.Close()
	log.Printf("SRR %t", ctrl.Done)
	d, err := ctrl.C.Read()
	for err == nil && !ctrl.Done {
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
				ctrl.sendToChan(rsp)
			case RespDig:
				var sd SendDig
				sd.Dig = v.Dig
				sd.HaveIt = v.HaveIt
				ctrl.sendToChan(sd)
			case RejectDig:
				err = ctrl.storeDataRejected(v.Dig)
				if err != nil {
					log.Printf("Failed to save rejected data record! %s", err)
				}
			case AckDig:
				err = ctrl.deleteSendData(v.Dig)
				if err != nil {
					log.Printf("Failed to delete SendData! %s", err)
				}
			case *gripdata.Node:
				err = IncomingNode(v, ctrl.DB)
				ctrl.processSendError("Node", v.Dig, err)
			case *gripdata.AssociateNodeAccountKey:
				err = IncomingNodeAccountKey(v, ctrl.DB)
				ctrl.processSendError("AssociateNodeAccountKey", v.Dig, err)
			case *gripdata.UseShareNodeKey:
				err = IncomingUseShareNodeKey(v, ctrl.DB)
				ctrl.processSendError("UseShareNodeKey", v.Dig, err)
			case *gripdata.ShareNodeInfo:
				err = IncomingShareNode(v, ctrl.DB)
				ctrl.processSendError("ShareNodeInfo", v.Dig, err)
			case *gripdata.Context:
				err = IncomingContext(v, ctrl.DB)
				ctrl.processSendError("Context", v.Dig, err)
			case *gripdata.ContextRequest:
				err = IncomingContextRequest(v, ctrl.DB)
				ctrl.processSendError("ContextRequest", v.Dig, err)
			case *gripdata.ContextResponse:
				err = IncomingContextResponse(v, ctrl.DB)
				ctrl.processSendError("ContextResponse", v.Dig, err)
			case *gripdata.ContextFile:
				err = IncomingContextFile(v, ctrl.DB)
				ctrl.processSendError("ContextFile", v.Dig, err)
			}
		}
		d, err = ctrl.C.Read()
		ctrl.LastReadLoop = uint64(time.Now().UnixNano())
	}
	if err != nil {
		log.Printf("Connection error: %s\n", err)
	}
}

func (ctrl *ConnectionController) setConnected(incomming bool) {
	nt := uint64(time.Now().UnixNano())
	myid, _ := ctrl.DB.GetPrivateNodeData()
	msid := base64.StdEncoding.EncodeToString(myid.ID)
	fsid := base64.StdEncoding.EncodeToString(ctrl.C.GetNodeID())
	log.Printf("New Connection %d, me: %s from: %s", ctrl.ConID, msid, fsid)
	err := ctrl.DB.SetNodeEphemeraConnected(incomming, ctrl.C.GetNodeID(), nt)
	if err != nil {
		log.Printf("Error setting connected: %s", err)
	}
}

func (ctrl *ConnectionController) setConnectionClosed() {
	log.Printf("Connection closed: %d", ctrl.ConID)
	err := ctrl.DB.SetNodeEphemeraClosed(ctrl.C.GetNodeID())
	if err != nil {
		log.Printf("Error setting connection closed: %s", err)
	}
}
