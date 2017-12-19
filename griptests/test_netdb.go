package griptests

import (
	"bytes"
	"encoding/base64"
	"log"
	"reflect"

	"github.com/wyathan/grip/gripdata"
)

/*
	StoreSendData(s *gripdata.SendData) error
	StoreRejectedSendData(s *gripdata.RejectedSendData) error
	GetSendData(target []byte, max int) []gripdata.SendData //get all send data for target node
	DeleteSendData(d []byte, to []byte) error //Data has been setnt to the node
	GetDigestData(d []byte) interface{}
	GetConnectableNodesWithSendData(max int, curtime uint64) []gripdata.NodeEphemera
	GetConnectableNodesWithShareNodeKey(max int, curtime uint64) []gripdata.NodeEphemera
	GetConnectableUseShareKeyNodes(max int, curtime uint64) []gripdata.NodeEphemera
	GetConnectableAny(max int, curtime uint64) []gripdata.NodeEphemera
	GetAllConnected() []gripdata.NodeEphemera
	CreateNodeEphemera(id []byte, connectable bool) error
	SetNodeEphemeraNextConnection(id []byte, last uint64, next uint64) error
	ClearAllConnected()
	CanNodeEphemeraGoPending(id []byte) bool
	SetNodeEphemeraConnected(incomming bool, id []byte, curtime uint64) error
	SetNodeEphemeraClosed(id []byte) error
*/

func (t *TestDB) StoreSendData(s *gripdata.SendData) error {
	t.Lock()
	defer t.Unlock()
	if t.SendData == nil {
		t.SendData = make(map[string][]gripdata.SendData)
	}
	tk := base64.StdEncoding.EncodeToString(s.TargetID)
	kl := t.SendData[tk]
	t.SendData[tk] = append(kl, *s) //convenient all send data will already be sorted
	return nil
}
func (t *TestDB) StoreRejectedSendData(s *gripdata.RejectedSendData) error {
	t.Lock()
	defer t.Unlock()
	nid := base64.StdEncoding.EncodeToString(s.TargetID)
	r := t.RejectedData[nid]
	t.RejectedData[nid] = append(r, *s)
	return nil
}
func (t *TestDB) GetSendData(target []byte, max int) []gripdata.SendData {
	t.Lock()
	defer t.Unlock()
	tk := base64.StdEncoding.EncodeToString(target)
	var r []gripdata.SendData
	rc := t.SendData[tk] //Already sorted
	for c := 0; c < len(rc) && c < max; c++ {
		r = append(r, rc[c])
	}
	return r
}
func (t *TestDB) DeleteSendData(d []byte, to []byte) error {
	t.Lock()
	defer t.Unlock()
	tk := base64.StdEncoding.EncodeToString(to)
	sl := t.SendData[tk]
	var nl []gripdata.SendData
	for _, v := range sl {
		if !bytes.Equal(v.Dig, d) {
			nl = append(nl, v)
		}
	}
	t.SendData[tk] = nl
	return nil
}
func (t *TestDB) GetDigestData(d []byte) interface{} {
	t.Lock()
	defer t.Unlock()
	ds := base64.StdEncoding.EncodeToString(d)
	v := t.DigData[ds]
	if v != nil {
		log.Printf("GetDigestData %s  %s\n", ds, reflect.TypeOf(v).String())
	} else {
		log.Printf("GetDigestData %s  nil\n", ds)
	}
	return v
}
func (t *TestDB) GetConnectableNodesWithSendData(max int, curtime uint64) []gripdata.NodeEphemera {
	t.Lock()
	defer t.Unlock()
	var r []gripdata.NodeEphemera
	added := make(map[string]bool)
	for _, v := range t.NodeEphemera {
		tk := base64.StdEncoding.EncodeToString(v.ID)
		if !v.Connected && len(r) < max && v.NextAttempt <= curtime && v.Connectable && !added[tk] {
			if len(t.SendData[tk]) > 0 {
				added[tk] = true
				r = append(r, *v)
			}
		}
	}
	return r
}
func (t *TestDB) GetConnectableNodesWithShareNodeKey(max int, curtime uint64) []gripdata.NodeEphemera {
	t.Lock()
	defer t.Unlock()
	ids := base64.StdEncoding.EncodeToString(t.LclPrvNodeData.ID)
	lk := t.ShareNodes[ids]
	var r []gripdata.NodeEphemera
	for _, sn := range lk {
		if sn.Key != "" {
			tk := base64.StdEncoding.EncodeToString(sn.TargetNodeID)
			v := t.NodeEphemera[tk]
			if v != nil {
				if !v.Connected && len(r) < max && v.NextAttempt <= curtime && v.Connectable {
					r = append(r, *v)
				}
			}
		}
	}
	return r
}
func (t *TestDB) GetConnectableUseShareKeyNodes(max int, curtime uint64) []gripdata.NodeEphemera {
	t.Lock()
	defer t.Unlock()
	var r []gripdata.NodeEphemera
	for _, snl := range t.UseShareKeys {
		for _, s := range snl {
			if bytes.Equal(s.NodeID, t.LclPrvNodeData.ID) {
				tk := base64.StdEncoding.EncodeToString(s.TargetID)
				v := t.NodeEphemera[tk]
				if v != nil {
					if !v.Connected && len(r) < max && v.NextAttempt <= curtime && v.Connectable {
						r = append(r, *v)
					}
				}
			}
		}
	}
	return r
}
func (t *TestDB) GetConnectableAny(max int, curtime uint64) []gripdata.NodeEphemera {
	t.Lock()
	defer t.Unlock()
	var r []gripdata.NodeEphemera
	for _, v := range t.NodeEphemera {
		if !v.Connected && len(r) < max && v.NextAttempt <= curtime && v.Connectable {
			r = append(r, *v)
		}
	}
	return r
}
func (t *TestDB) GetAllConnected() []gripdata.NodeEphemera {
	t.Lock()
	defer t.Unlock()
	var r []gripdata.NodeEphemera
	for _, v := range t.NodeEphemera {
		if v.Connected {
			r = append(r, *v)
		}
	}
	return r
}
func (t *TestDB) CreateNodeEphemera(id []byte, connectable bool) error {
	t.Lock()
	defer t.Unlock()
	tk := base64.StdEncoding.EncodeToString(id)
	ep := t.NodeEphemera[tk]
	if ep == nil {
		var nep gripdata.NodeEphemera
		nep.ID = id
		ep = &nep
		t.NodeEphemera[tk] = ep
	}
	ep.Connectable = connectable
	return nil
}
func (t *TestDB) SetNodeEphemeraNextConnection(id []byte, last uint64, next uint64) error {
	t.Lock()
	defer t.Unlock()
	tk := base64.StdEncoding.EncodeToString(id)
	ep := t.NodeEphemera[tk]
	if ep != nil {
		ep.LastConnAttempt = last
		ep.NextAttempt = next
		ep.ConnectionPending = false
	}
	return nil
}
func (t *TestDB) ClearAllConnected() {
	t.Lock()
	defer t.Unlock()
	for _, v := range t.NodeEphemera {
		v.Connected = false
	}
}
func (t *TestDB) CanNodeEphemeraGoPending(id []byte) bool {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(id)
	nid := t.NodeEphemera[sid]
	if nid == nil {
		var n gripdata.NodeEphemera
		n.ID = id
		nid = &n
		t.NodeEphemera[sid] = nid
	}
	if !nid.Connected && !nid.ConnectionPending {
		nid.ConnectionPending = true
		return true
	}
	return false
}
func (t *TestDB) SetNodeEphemeraConnected(incomming bool, id []byte, curtime uint64) error {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(id)
	nid := t.NodeEphemera[sid]
	if nid == nil {
		var n gripdata.NodeEphemera
		n.ID = id
		nid = &n
		t.NodeEphemera[sid] = nid
	}
	nid.Connected = true
	nid.ConnectionPending = false
	if incomming {
		nid.LastConnReceived = curtime
	} else {
		nid.LastConnection = curtime
	}
	return nil
}
func (t *TestDB) SetNodeEphemeraClosed(id []byte) error {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(id)
	nid := t.NodeEphemera[sid]
	if nid != nil {
		nid.Connected = false
		nid.ConnectionPending = false
	}
	return nil
}
