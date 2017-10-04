package griptests

import (
	"bytes"
	"encoding/base64"
	"log"
	"reflect"
	"sync"

	"github.com/wyathan/grip"
	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

type TestDB struct {
	sync.Mutex
	LclNode        *gripdata.Node
	LclPrvNodeData *gripdata.MyNodePrivateData
	Accounts       map[string]*gripdata.Account
	AccountKeys    map[string]*gripdata.NodeAccountKey
	NodeAccounts   map[string]*gripdata.NodeAccount
	Nodes          map[string]*gripdata.Node
	ShareNodes     map[string][]gripdata.ShareNodeInfo
	ShareNodeKeys  map[string][]gripdata.ShareNodeInfo
	UseShareKeys   map[string][]gripdata.UseShareNodeKey
	SendData       map[string][]gripdata.SendData
	DigData        map[string]interface{}
	NodeEphemera   map[string]*gripdata.NodeEphemera
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
		if !bytes.Equal(v.Dig, to) {
			nl = append(nl, v)
		}
	}
	t.SendData[tk] = nl
	return nil
}
func (t *TestDB) GetConnectableNodesWithSendData(max int, curtime uint64) []gripdata.NodeEphemera {
	t.Lock()
	defer t.Unlock()
	var r []gripdata.NodeEphemera
	for _, v := range t.NodeEphemera {
		if !v.Connected && len(r) < max && v.NextAttempt <= curtime {
			tk := base64.StdEncoding.EncodeToString(v.ID)
			if len(t.SendData[tk]) > 0 {
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
func (t *TestDB) GetNodeEphemera(id []byte) *gripdata.NodeEphemera {
	t.Lock()
	defer t.Unlock()
	tk := base64.StdEncoding.EncodeToString(id)
	return t.NodeEphemera[tk]
}
func (t *TestDB) StoreNodeEphemera(ne *gripdata.NodeEphemera) error {
	t.Lock()
	defer t.Unlock()
	tk := base64.StdEncoding.EncodeToString(ne.ID)
	t.NodeEphemera[tk] = ne
	return nil
}

func testTestDBImplements() {
	var t TestDB
	var s grip.NodeNetAccountdb
	s = &t
	s.GetAccount("")
}
