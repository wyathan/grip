package grip

import (
	"log"
	"reflect"
	"time"

	"github.com/wyathan/grip/gripdata"
)

func (ctrl *ConnectionController) sendIntroduction() {
	myn, _ := ctrl.DB.GetPrivateNodeData()
	err := ctrl.C.Send(myn)
	if err != nil {
		log.Print("Failed to send my node data for connection init")
		ctrl.Close()
	}
}

func (ctrl *ConnectionController) sendTimeout(sent int) <-chan bool {
	timeout := make(chan bool, 1)
	go func(ct int) {
		if ct == 0 {
			time.Sleep(SLEEPONNOSEND)
		}
		timeout <- true
	}(sent)
	return timeout
}

func (ctrl *ConnectionController) sendData(r interface{}) (int, error) {
	var err error
	sent := 0
	switch v := r.(type) {
	case SendDig:
		if v.HaveIt {
			ctrl.deleteSendData(v.Dig)
		} else {
			err = ctrl.sendSendData(v.Dig)
		}
	case bool:
		sent, err = ctrl.sendFromDatabase()
	default:
		err = ctrl.C.Send(v)
	}
	return sent, err
}

func (ctrl *ConnectionController) sendLoop(sent int) {
	for !ctrl.Done {
		var err error
		sent := 0
		timeout := ctrl.sendTimeout(sent)
		select {
		case r, ok := <-ctrl.SendChan:
			if !ok || ctrl.Done {
				//Do not call ctrl.Close() because
				//its only purpose is to close ctlr.S
				//so that this routine exits.
				ctrl.Done = true
			} else {
				sent, err = ctrl.sendData(r)
			}
		case <-timeout:
			sent, err = ctrl.sendFromDatabase()
		}
		if err != nil {
			ctrl.Close()
		}
		ctrl.LastWriteLoop = uint64(time.Now().UnixNano())
	}
}

//ConnectionWriteRoutine go routine to write to connections
func (ctrl *ConnectionController) ConnectionWriteRoutine() {
	defer func() {
		ctrl.setConnectionClosed()
		ctrl.C.Close() //Only close connection here
	}()
	if !ctrl.Done {
		ctrl.setConnected(ctrl.Incoming)
		ctrl.sendIntroduction()
		sent, err := ctrl.sendFromDatabase()
		if err != nil {
			ctrl.Close()
		}
		ctrl.sendLoop(sent)
	}
}

func (ctrl *ConnectionController) readSwitch(d interface{}) (err error) {
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
		err = ctrl.dataRejected(v.Dig)
		if err != nil {
			log.Printf("Failed to process rejection! %s", err)
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
	return err
}

func (ctrl *ConnectionController) readLoop() error {
	d, err := ctrl.C.Read()
	for err == nil && !ctrl.Done {
		if d != nil {
			err = ctrl.readSwitch(d)
			d, err = ctrl.C.Read()
			ctrl.LastReadLoop = uint64(time.Now().UnixNano())
		}
	}
	return err
}

//ConnectionReadRoutine go routine to read from connections
func (ctrl *ConnectionController) ConnectionReadRoutine() {
	defer ctrl.Close()
	err := ctrl.readLoop()
	if err != nil {
		log.Printf("Connection error: %s\n", err)
	}
}
