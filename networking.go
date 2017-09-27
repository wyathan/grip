package grip

import (
	"log"
	"sync"
	"time"

	"github.com/wyathan/grip/gripdata"
)

//Connection handles sending and reading data
//on a connection
type Connection interface {
	Read() (interface{}, error)
	Send(d interface{}) error
	Close()
}

//Socket handles making connections and accepting them
type Socket interface {
	ConnectTo(n *gripdata.Node) (Connection, error)
	Accept() (Connection, error)
}

//ConnectionController is a handle for connections
type ConnectionController struct {
	sync.Mutex
	ID   []byte //ID of node connected to
	C    Connection
	S    chan *gripdata.SendData
	Done bool
	DB   NodeNetAccountdb
}

//Close a connection
func (ctrl *ConnectionController) Close() {
	ctrl.Lock()
	defer ctrl.Unlock()
	if !ctrl.Done {
		ctrl.Done = true
		close(ctrl.S)
	}
}

func (ctrl *ConnectionController) sendFromDatabase() error {
	c := 0
	if ctrl.ID != nil {
		sl := ctrl.DB.GetSendData(ctrl.ID, 1000)
		err := ctrl.sendSendDataList(sl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *ConnectionController) sendSendDataList(sl []gripdata.SendData) error {
	for _, v := range sl {
		err := ctrl.sendSendData(&v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *ConnectionController) sendSendData(d *gripdata.SendData) error {
	sd := ctrl.DB.GetDigestData(d.Dig)
	if sd != nil {
		err := ctrl.C.Send(sd)
		if err != nil {
			return err
		}
		ctrl.DB.DeleteSendData(d)
	}
	return nil
}

//ConnectionWriteRoutine go routine to write to connections
func (ctrl *ConnectionController) ConnectionWriteRoutine() {
	defer ctrl.C.Close() //Only close connection here
	c, err := ctrl.sendFromDatabase()
	for err == nil && !ctrl.Done {

		timeout := make(chan bool, 1)
		go func(ct int) {
			if ct == 0 {
				time.Sleep(10 * time.Second)
			}
			timeout <- true
		}(c)
		select {
		case d, ok := <-ctrl.S:
			if !ok || ctrl.Done {
				ctrl.Done = true
			} else {
				ctrl.sendSendData(d)
			}
		case <-timeout:
			ctrl.sendFromDatabase()
		}
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
	log.Println(err)
}
