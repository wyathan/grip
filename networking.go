package grip

import (
	"encoding/base64"
	"log"
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
}

//SocketController handles a socket
type SocketController struct {
	sync.Mutex
	Connections map[string]*ConnectionController
	S           Socket
	Done        bool
	DB          NodeNetAccountdb
}

func (s *SocketController) addConnection(c *ConnectionController) bool {
	s.Lock()
	defer s.Unlock()
	ck := base64.StdEncoding.EncodeToString(c.C.GetNodeID())
	cc := s.Connections[ck]
	if cc != nil {
		c.SocketCtrl = nil
		c.Close()
		return false
	}
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

func (s *SocketController) listConnections() []*ConnectionController {
	s.Lock()
	defer s.Unlock()
	var r []*ConnectionController
	for _, v := range s.Connections {
		r = append(r, v)
	}
	return r
}

//ListenRoutine handles new incoming connections
func (s *SocketController) ListenRoutine() {
	con, err := s.S.Accept()
	for !s.Done && err == nil {
		if con.GetNodeID() == nil {
			log.Printf("This is bad.  Connections should only return with valid ID")
		} else {
			var ctrl ConnectionController
			ctrl.C = con
			ctrl.DB = s.DB
			ctrl.SendChan = make(chan interface{}, BUFFERSIZE)
			ctrl.SocketCtrl = s
			s.addConnection(&ctrl)
			go ctrl.ConnectionReadRoutine()
			go ctrl.ConnectionWriteRoutine()
			con, err = s.S.Accept()
		}
	}
}

//ConnectionController is a handle for connections
type ConnectionController struct {
	sync.Mutex
	C          Connection
	SendChan   chan interface{}
	Done       bool
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
}

//ConnectionReadRoutine go routine to read from connections
func (ctrl *ConnectionController) ConnectionReadRoutine() {
	defer ctrl.Close()
	d, err := ctrl.C.Read()
	for err == nil {
		if d != nil {
			switch v := d.(type) {
			default:
				log.Printf("Unknown type: %s\n", v)
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
			case gripdata.Node:
				log.Println("Node data received")
				err = IncomingNode(&v, ctrl.DB)
			case gripdata.AssociateNodeAccountKey:
				log.Println("AssociateNodeAccountKey data received")
				err = IncomingNodeAccountKey(&v, ctrl.DB)
			case gripdata.UseShareNodeKey:
				log.Println("UseShareNodeKey data received")
				err = IncomingUseShareNodeKey(&v, ctrl.DB)
			case gripdata.ShareNodeInfo:
				log.Println("ShareNodeInfo data received")
				err = IncomingShareNode(&v, ctrl.DB)
			}
		}
		d, err = ctrl.C.Read()
	}
	if err != nil {
		log.Printf("Connection error: %s\n", err)
	}
}
