package griptests

import (
	"encoding/base64"

	"github.com/wyathan/grip/gripdata"
)

/*
	GetNode(id []byte) *gripdata.Node
	GetPrivateNodeData() (*gripdata.Node, *gripdata.MyNodePrivateData)
	StoreNode(n *gripdata.Node) error
	StoreMyPrivateNodeData(n *gripdata.Node, pr *gripdata.MyNodePrivateData) error
	StoreShareNodeInfo(sn *gripdata.ShareNodeInfo) error
	ListNodes() []gripdata.Node                           //List all known nodes
	ListShareNodeInfo(id []byte) []gripdata.ShareNodeInfo //list all ShareNodeInfo from node
	ListShareNodeKey(s string) []gripdata.ShareNodeInfo   //List all ShareNodeInfo with key
	StoreUseShareNodeKey(k *gripdata.UseShareNodeKey) error
	ListUseShareNodeKey(k string) []gripdata.UseShareNodeKey //List all that have used key
	StoreAssociateNodeAccountKey(n *gripdata.AssociateNodeAccountKey) error
*/

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
func (t *TestDB) StoreAssociateNodeAccountKey(c *gripdata.AssociateNodeAccountKey) error {
	t.Lock()
	defer t.Unlock()
	t.addDig(c)
	return nil
}
