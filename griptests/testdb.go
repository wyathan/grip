package griptests

import (
	"bytes"
	"encoding/base64"
	"errors"
	"log"
	"reflect"
	"sync"

	"github.com/wyathan/grip"
	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

//TestDB is that
type TestDB struct {
	sync.Mutex
	LclNode             *gripdata.Node
	LclPrvNodeData      *gripdata.MyNodePrivateData
	Accounts            map[string]*gripdata.Account
	AccountKeys         map[string]*gripdata.NodeAccountKey
	NodeAccounts        map[string]*gripdata.NodeAccount
	Nodes               map[string]*gripdata.Node
	ShareNodes          map[string][]gripdata.ShareNodeInfo
	ShareNodeKeys       map[string][]gripdata.ShareNodeInfo
	UseShareKeys        map[string][]gripdata.UseShareNodeKey
	SendData            map[string][]gripdata.SendData
	DigData             map[string]interface{}
	NodeEphemera        map[string]*gripdata.NodeEphemera
	Contexts            map[string]*gripdata.Context
	ContextRequests     map[string][]gripdata.ContextRequest
	ContextResponses    map[string]map[string]*gripdata.ContextResponse
	ContextFiles        map[string][]gripdata.ContextFile
	RejectedData        map[string][]gripdata.RejectedSendData
	VeryBadContextFiles []gripdata.ContextFile
}

func NewTestDB() *TestDB {
	var t TestDB
	t.Accounts = make(map[string]*gripdata.Account)
	t.AccountKeys = make(map[string]*gripdata.NodeAccountKey)
	t.NodeAccounts = make(map[string]*gripdata.NodeAccount)
	t.Nodes = make(map[string]*gripdata.Node)
	t.ShareNodes = make(map[string][]gripdata.ShareNodeInfo)
	t.ShareNodeKeys = make(map[string][]gripdata.ShareNodeInfo)
	t.UseShareKeys = make(map[string][]gripdata.UseShareNodeKey)
	t.SendData = make(map[string][]gripdata.SendData)
	t.DigData = make(map[string]interface{})
	t.NodeEphemera = make(map[string]*gripdata.NodeEphemera)
	t.Contexts = make(map[string]*gripdata.Context)
	t.ContextRequests = make(map[string][]gripdata.ContextRequest)
	t.ContextResponses = make(map[string]map[string]*gripdata.ContextResponse)
	t.ContextFiles = make(map[string][]gripdata.ContextFile)
	t.RejectedData = make(map[string][]gripdata.RejectedSendData)

	return &t
}

func (t *TestDB) addDig(s gripcrypto.SignInf) {
	if t.DigData == nil {
		t.DigData = make(map[string]interface{})
	}
	ds := base64.StdEncoding.EncodeToString(s.GetDig())
	log.Printf("ADDDIG %s %s\n", ds, reflect.TypeOf(s).String())
	t.DigData[ds] = s
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

func (t *TestDB) StoreAccount(a *gripdata.Account) error {
	t.Lock()
	defer t.Unlock()
	if t.Accounts == nil {
		t.Accounts = make(map[string]*gripdata.Account)
	}
	t.Accounts[a.AccountID] = a
	return nil
}
func (t *TestDB) IncrNumberContexts(a *gripdata.Account) error {
	t.Lock()
	defer t.Unlock()
	ua := t.Accounts[a.AccountID]
	ua.NumberContexts++
	return nil
}
func (t *TestDB) IncrNumberNodes(a *gripdata.Account) error {
	t.Lock()
	defer t.Unlock()
	ua := t.Accounts[a.AccountID]
	ua.NumberNodes++
	return nil
}
func (t *TestDB) GetAccount(s string) *gripdata.Account {
	t.Lock()
	defer t.Unlock()
	return t.Accounts[s]
}
func (t *TestDB) StoreNodeAccountKey(ak *gripdata.NodeAccountKey) error {
	t.Lock()
	defer t.Unlock()
	if t.AccountKeys == nil {
		t.AccountKeys = make(map[string]*gripdata.NodeAccountKey)
	}
	t.AccountKeys[ak.Key] = ak
	return nil
}
func (t *TestDB) GetNodeAccountKey(key string) *gripdata.NodeAccountKey {
	t.Lock()
	defer t.Unlock()
	return t.AccountKeys[key]
}
func (t *TestDB) StoreNodeAccount(na *gripdata.NodeAccount) error {
	t.Lock()
	defer t.Unlock()
	if t.NodeAccounts == nil {
		t.NodeAccounts = make(map[string]*gripdata.NodeAccount)
	}
	t.NodeAccounts[base64.StdEncoding.EncodeToString(na.NodeID)] = na
	return nil
}
func (t *TestDB) GetNodeAccount(id []byte) *gripdata.NodeAccount {
	t.Lock()
	defer t.Unlock()
	return t.NodeAccounts[base64.StdEncoding.EncodeToString(id)]
}

func (t *TestDB) GetNode(id []byte) *gripdata.Node {
	t.Lock()
	defer t.Unlock()
	return t.Nodes[base64.StdEncoding.EncodeToString(id)]
}
func (t *TestDB) GetPrivateNodeData() (*gripdata.Node, *gripdata.MyNodePrivateData) {
	t.Lock()
	defer t.Unlock()
	return t.LclNode, t.LclPrvNodeData
}
func (t *TestDB) StoreNode(n *gripdata.Node) error {
	t.Lock()
	defer t.Unlock()
	if t.Nodes == nil {
		t.Nodes = make(map[string]*gripdata.Node)
	}
	t.Nodes[base64.StdEncoding.EncodeToString(n.ID)] = n
	t.addDig(n)
	return nil
}
func (t *TestDB) StoreMyPrivateNodeData(n *gripdata.Node, pr *gripdata.MyNodePrivateData) error {
	t.Lock()
	defer t.Unlock()
	t.LclNode = n
	t.LclPrvNodeData = pr
	t.addDig(n)
	t.Nodes[base64.StdEncoding.EncodeToString(n.ID)] = n
	return nil
}
func (t *TestDB) StoreShareNodeInfo(sn *gripdata.ShareNodeInfo) error {
	t.Lock()
	defer t.Unlock()
	if t.ShareNodes == nil {
		t.ShareNodes = make(map[string][]gripdata.ShareNodeInfo)
	}
	ids := base64.StdEncoding.EncodeToString(sn.NodeID)
	lk := t.ShareNodes[ids]
	t.ShareNodes[ids] = append(lk, *sn)
	if sn.Key != "" {
		if t.ShareNodeKeys == nil {
			t.ShareNodeKeys = make(map[string][]gripdata.ShareNodeInfo)
		}
		lks := t.ShareNodeKeys[sn.Key]
		t.ShareNodeKeys[sn.Key] = append(lks, *sn)
	}
	t.addDig(sn)
	return nil
}
func (t *TestDB) ListNodes() []gripdata.Node {
	t.Lock()
	defer t.Unlock()
	var r []gripdata.Node
	for _, v := range t.Nodes {
		r = append(r, *v)
	}
	return r
}
func (t *TestDB) ListShareNodeInfo(id []byte) []gripdata.ShareNodeInfo {
	t.Lock()
	defer t.Unlock()
	return t.ShareNodes[base64.StdEncoding.EncodeToString(id)]
}
func (t *TestDB) ListShareNodeKey(s string) []gripdata.ShareNodeInfo {
	t.Lock()
	defer t.Unlock()
	return t.ShareNodeKeys[s]
}
func (t *TestDB) StoreUseShareNodeKey(k *gripdata.UseShareNodeKey) error {
	t.Lock()
	defer t.Unlock()
	if t.UseShareKeys == nil {
		t.UseShareKeys = make(map[string][]gripdata.UseShareNodeKey)
	}
	kl := t.UseShareKeys[k.Key]
	t.UseShareKeys[k.Key] = append(kl, *k)
	t.addDig(k)
	return nil
}
func (t *TestDB) ListUseShareNodeKey(k string) []gripdata.UseShareNodeKey {
	t.Lock()
	defer t.Unlock()
	return t.UseShareKeys[k]
}

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
	}
	return nil
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
func (t *TestDB) StoreContext(c *gripdata.Context) error {
	t.Lock()
	defer t.Unlock()
	id := base64.StdEncoding.EncodeToString(c.Dig)
	t.Contexts[id] = c
	t.addDig(c)
	return nil
}
func (t *TestDB) GetContext(id []byte) *gripdata.Context {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(id)
	return t.Contexts[sid]
}
func (t *TestDB) GetContextRequests(id []byte) []gripdata.ContextRequest {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(id)
	return t.ContextRequests[sid]
}
func (t *TestDB) StoreContextRequest(c *gripdata.ContextRequest) error {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(c.Dig)
	t.ContextRequests[sid] = append(t.ContextRequests[sid], *c)
	t.addDig(c)
	return nil
}
func (t *TestDB) GetContextRequest(cid []byte, tgtid []byte) *gripdata.ContextRequest {
	t.Lock()
	defer t.Unlock()
	scid := base64.StdEncoding.EncodeToString(cid)
	tar := t.ContextRequests[scid]
	for _, cr := range tar {
		if bytes.Equal(tgtid, cr.TargetNodeID) {
			return &cr
		}
	}
	return nil
}
func (t *TestDB) StoreContextResponse(c *gripdata.ContextResponse) error {
	t.Lock()
	defer t.Unlock()
	cid := base64.StdEncoding.EncodeToString(c.ContextDig)
	tid := base64.StdEncoding.EncodeToString(c.TargetNodeID)
	m := t.ContextResponses[cid]
	if m == nil {
		m = make(map[string]*gripdata.ContextResponse)
		t.ContextResponses[cid] = m
	}
	m[tid] = c
	t.addDig(c)
	return nil
}

func (t *TestDB) GetContextResponse(cid []byte, tgtid []byte) *gripdata.ContextResponse {
	t.Lock()
	defer t.Unlock()
	scid := base64.StdEncoding.EncodeToString(cid)
	stid := base64.StdEncoding.EncodeToString(tgtid)
	rsp := t.ContextResponses[scid]
	if rsp != nil {
		return rsp[stid]
	}
	return nil
}

func (t *TestDB) GetContextFileByDepDataDig(d []byte) *gripdata.ContextFile {
	t.Lock()
	defer t.Unlock()
	return nil
}

func (t *TestDB) GetContextResponses(cid []byte) []gripdata.ContextResponse {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(cid)
	rm := t.ContextResponses[sid]
	var r []gripdata.ContextResponse
	for _, v := range rm {
		r = append(r, *v)
	}
	return nil
}

func (t *TestDB) StoreContextFile(c *gripdata.ContextFile) error {
	t.Lock()
	defer t.Unlock()
	cid := base64.StdEncoding.EncodeToString(c.Context)
	fl := t.ContextFiles[cid]
	t.ContextFiles[cid] = append(fl, *c)
	t.addDig(c)
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

func (t *TestDB) StoreVeryBadContextFile(c *gripdata.ContextFile) error {
	t.Lock()
	defer t.Unlock()
	t.VeryBadContextFiles = append(t.VeryBadContextFiles, *c)
	return nil
}

func (t *TestDB) CheckUpdateStorageUsed(a *gripdata.Account, fsize uint64) error {
	t.Lock()
	defer t.Unlock()
	a = t.Accounts[a.AccountID]
	if a.DiskSpaceUsed+fsize > a.MaxDiskSpace {
		return errors.New("would exceed max storage")
	}
	a.DiskSpaceUsed = a.DiskSpaceUsed + fsize
	t.Accounts[a.AccountID] = a
	return nil
}

//testTestDBImplements make sure we implement the interfaces
func testTestDBImplements() {
	var t TestDB
	var s grip.NodeNetAccountContextdb
	s = &t
	s.GetAccount("")
}
