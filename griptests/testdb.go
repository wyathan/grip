package griptests

import (
	"bytes"
	"encoding/base64"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

type TestDB struct {
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
}

func (t *TestDB) addDig(s gripcrypto.SignInf) {
	if t.DigData == nil {
		t.DigData = make(map[string]interface{})
	}
	ds := base64.StdEncoding.EncodeToString(s.GetDig())
	t.DigData[ds] = s
}

func (t *TestDB) GetDigestData(d []byte) interface{} {
	ds := base64.StdEncoding.EncodeToString(d)
	return t.DigData[ds]
}

func (t *TestDB) StoreAccount(a *gripdata.Account) error {
	if t.Accounts == nil {
		t.Accounts = make(map[string]*gripdata.Account)
	}
	t.Accounts[a.AccountID] = a
	return nil
}
func (t *TestDB) GetAccount(s string) *gripdata.Account {
	return t.Accounts[s]
}
func (t *TestDB) StoreNodeAccountKey(ak *gripdata.NodeAccountKey) error {
	if t.AccountKeys == nil {
		t.AccountKeys = make(map[string]*gripdata.NodeAccountKey)
	}
	t.AccountKeys[ak.Key] = ak
	return nil
}
func (t *TestDB) GetNodeAccountKey(key string) *gripdata.NodeAccountKey {
	return t.AccountKeys[key]
}
func (t *TestDB) StoreNodeAccount(na *gripdata.NodeAccount) error {
	if t.NodeAccounts == nil {
		t.NodeAccounts = make(map[string]*gripdata.NodeAccount)
	}
	t.NodeAccounts[base64.StdEncoding.EncodeToString(na.NodeID)] = na
	return nil
}
func (t *TestDB) GetNodeAccount(id []byte) *gripdata.NodeAccount {
	return t.NodeAccounts[base64.StdEncoding.EncodeToString(id)]
}

func (t *TestDB) GetNode(id []byte) *gripdata.Node {
	return t.Nodes[base64.StdEncoding.EncodeToString(id)]
}
func (t *TestDB) GetPrivateNodeData() (*gripdata.Node, *gripdata.MyNodePrivateData) {
	return t.LclNode, t.LclPrvNodeData
}
func (t *TestDB) StoreNode(n *gripdata.Node) error {
	if t.Nodes == nil {
		t.Nodes = make(map[string]*gripdata.Node)
	}
	t.Nodes[base64.StdEncoding.EncodeToString(n.ID)] = n
	t.addDig(n)
	return nil
}
func (t *TestDB) StoreMyPrivateNodeData(n *gripdata.Node, pr *gripdata.MyNodePrivateData) error {
	t.LclNode = n
	t.LclPrvNodeData = pr
	t.addDig(n)
	return nil
}
func (t *TestDB) StoreShareNodeInfo(sn *gripdata.ShareNodeInfo) error {
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
	var r []gripdata.Node
	for _, v := range t.Nodes {
		r = append(r, *v)
	}
	return r
}
func (t *TestDB) ListShareNodeInfo(id []byte) []gripdata.ShareNodeInfo {
	return t.ShareNodes[base64.StdEncoding.EncodeToString(id)]
}
func (t *TestDB) ListShareNodeKey(s string) []gripdata.ShareNodeInfo {
	return t.ShareNodeKeys[s]
}
func (t *TestDB) StoreUseShareNodeKey(k *gripdata.UseShareNodeKey) error {
	if t.UseShareKeys == nil {
		t.UseShareKeys = make(map[string][]gripdata.UseShareNodeKey)
	}
	kl := t.UseShareKeys[k.Key]
	t.UseShareKeys[k.Key] = append(kl, *k)
	t.addDig(k)
	return nil
}
func (t *TestDB) ListUseShareNodeKey(k string) []gripdata.UseShareNodeKey {
	return t.UseShareKeys[k]
}

func (t *TestDB) StoreSendData(s *gripdata.SendData) error {
	if t.SendData == nil {
		t.SendData = make(map[string][]gripdata.SendData)
	}
	tk := base64.StdEncoding.EncodeToString(s.TargetID)
	kl := t.SendData[tk]
	t.SendData[tk] = append(kl, *s) //convenient all send data will already be sorted
	return nil
}
func (t *TestDB) GetSendData(target []byte, max int) []gripdata.SendData {
	tk := base64.StdEncoding.EncodeToString(target)
	return t.SendData[tk] //Already sorted
}
func (t *TestDB) DeleteSendData(s *gripdata.SendData) error {
	tk := base64.StdEncoding.EncodeToString(s.TargetID)
	sl := t.SendData[tk]
	var nl []gripdata.SendData
	for _, v := range sl {
		if !bytes.Equal(v.Dig, s.Dig) {
			nl = append(nl, v)
		}
	}
	t.SendData[tk] = nl
	return nil
}
