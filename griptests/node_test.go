package griptests

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/wyathan/grip"

	"github.com/wyathan/grip/gripdata"
)

type TestNodeDb struct {
}

func (a *TestNodeDb) ListNodes() []gripdata.Node {
	return make([]gripdata.Node, 0)
}

func (a *TestNodeDb) GetNode(id []byte) *gripdata.Node {
	return nil
}
func (a *TestNodeDb) GetPrivateNodeData() (*gripdata.Node, *gripdata.MyNodePrivateData) {
	return nil, nil
}
func (a *TestNodeDb) GetShareNodeInfo(id []byte, target []byte) *gripdata.ShareNodeInfo {
	return nil
}
func (a *TestNodeDb) StoreNode(n *gripdata.Node) error {
	return nil
}
func (a *TestNodeDb) StoreMyPrivateNodeData(n *gripdata.Node, pr *gripdata.MyNodePrivateData) error {
	return nil
}
func (a *TestNodeDb) StoreShareNodeInfo(sn *gripdata.ShareNodeInfo) error {
	return nil
}
func (a *TestNodeDb) ListShareNodeInfo(id []byte) []gripdata.ShareNodeInfo {
	return make([]gripdata.ShareNodeInfo, 0)
}
func (a *TestNodeDb) ListShareNodeKey(s string) []gripdata.ShareNodeInfo {
	return make([]gripdata.ShareNodeInfo, 0)
}
func (a *TestNodeDb) StoreUseShareNodeKey(k *gripdata.UseShareNodeKey) error {
	return nil
}
func (a *TestNodeDb) ListUseShareNodeKey(k string) []gripdata.UseShareNodeKey {
	return make([]gripdata.UseShareNodeKey, 0)
}
func (a *TestNodeDb) StoreAssociateNodeAccountKey(n *gripdata.AssociateNodeAccountKey) error {
	return nil
}

func TestNodeCreateVerify(t *testing.T) {
	var db TestNodeDb
	var n gripdata.Node
	var pr gripdata.MyNodePrivateData
	n.Connectable = false
	n.Name = "Test this node"
	n.URL = "over.there"
	pr.BindAddress = "right.here"
	pr.BindPort = 1111
	pr.PrivateMetaData = "some metadata"
	err := grip.CreateNewNode(&pr, &n, &db)
	if err != nil {
		t.Error(err)
	}
	err2 := grip.VerifyNode(&n)
	if err2 != nil {
		t.Error(err2)
	}
	fmt.Println(hex.EncodeToString(n.ID))
	fmt.Println(hex.EncodeToString(pr.ID))
}
