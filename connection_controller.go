package grip

import (
	"encoding/base64"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/orcaman/concurrent-map"
	"github.com/wyathan/grip/gripdata"
)

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

func (ctrl *ConnectionController) sendFromList(v *gripdata.SendData) error {
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
	return nil
}

func (ctrl *ConnectionController) sendSendDataList(sl []gripdata.SendData) error {
	for _, v := range sl {
		if ctrl.Done {
			return errors.New("Connection closed")
		}
		err := ctrl.sendFromList(&v)
		if err != nil {
			return err
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
	return ctrl.DB.DeleteSendData(d, ctrl.C.GetNodeID())
}

func (ctrl *ConnectionController) sendDataRejected(d []byte, msg string) error {
	var r RejectDig
	r.Dig = d
	r.Message = msg
	return ctrl.C.Send(r)
}

func (ctrl *ConnectionController) dataRejected(d []byte) error {
	err := ctrl.storeDataRejected(d)
	if err != nil {
		return err
	}
	err = ctrl.deleteSendData(d)
	if err != nil {
		return err
	}
	return nil
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

func (ctrl *ConnectionController) processSendError(fname string, dig []byte, err error) {
	if err == nil {
		log.Printf("%s data received", fname)
		ctrl.sendToChan(AckDig{Dig: dig})
	} else {
		log.Printf("%s data rejected: %s", fname, err)
		ctrl.sendDataRejected(dig, err.Error())
	}
}

func (ctrl *ConnectionController) setConnected(incomming bool) {
	nt := uint64(time.Now().UnixNano())
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
