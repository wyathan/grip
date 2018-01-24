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
func (a *TestNodeDb) StoreSendData(s *gripdata.SendData) error {
	return nil
}
func (a *TestNodeDb) StoreRejectedSendData(s *gripdata.RejectedSendData) error {
	return nil
}
func (a *TestNodeDb) GetSendData(target []byte, max int) []gripdata.SendData {
	return nil
}
func (a *TestNodeDb) DeleteSendData(d []byte, to []byte) (bool, error) {
	return false, nil
}
func (a *TestNodeDb) GetDigestData(d []byte) interface{} {
	return nil
}
func (a *TestNodeDb) GetConnectableNodesWithSendData(max int, curtime uint64) []gripdata.NodeEphemera {
	return nil
}
func (a *TestNodeDb) GetConnectableNodesWithShareNodeKey(max int, curtime uint64) []gripdata.NodeEphemera {
	return nil
}
func (a *TestNodeDb) GetConnectableUseShareKeyNodes(max int, curtime uint64) []gripdata.NodeEphemera {
	return nil
}
func (a *TestNodeDb) GetConnectableAny(max int, curtime uint64) []gripdata.NodeEphemera {
	return nil
}
func (a *TestNodeDb) GetAllConnected() []gripdata.NodeEphemera {
	return nil
}
func (a *TestNodeDb) CreateNodeEphemera(id []byte, connectable bool) error {
	return nil
}
func (a *TestNodeDb) SetNodeEphemeraNextConnection(id []byte, last uint64, next uint64) error {
	return nil
}
func (a *TestNodeDb) ClearAllConnected() {
}
func (a *TestNodeDb) CanNodeEphemeraGoPending(id []byte) bool {
	return false
}
func (a *TestNodeDb) SetNodeEphemeraConnected(incomming bool, id []byte, curtime uint64) error {
	return nil
}
func (a *TestNodeDb) SetNodeEphemeraClosed(id []byte) error {
	return nil
}
func (a *TestNodeDb) StoreContext(c *gripdata.Context) error {
	return nil
}
func (a *TestNodeDb) GetContext(id []byte) *gripdata.Context {
	return nil
}
func (a *TestNodeDb) GetContextRequests(id []byte) []*gripdata.ContextRequest {
	return nil
}
func (a *TestNodeDb) StoreContextRequest(c *gripdata.ContextRequest) error {
	return nil
}
func (a *TestNodeDb) GetContextRequest(cid []byte, tgtid []byte) *gripdata.ContextRequest {
	return nil
}
func (a *TestNodeDb) StoreContextResponse(c *gripdata.ContextResponse) error {
	return nil
}
func (a *TestNodeDb) GetContextResponse(cid []byte, tgtid []byte) *gripdata.ContextResponse {
	return nil
}
func (a *TestNodeDb) GetContextResponses(cid []byte) []*gripdata.ContextResponse {
	return nil
}
func (a *TestNodeDb) GetContextFileByDepDataDig(d []byte) *gripdata.ContextFileWrap {
	return nil
}
func (a *TestNodeDb) StoreVeryBadContextFile(c *gripdata.ContextFile) error {
	return nil
}
func (a *TestNodeDb) StoreContextFile(c *gripdata.ContextFile) (*gripdata.ContextFileWrap, error) {
	return nil, nil
}
func (a *TestNodeDb) GetAllThatDependOn(cid []byte, dig []byte) []*gripdata.ContextFileWrap {
	return nil
}
func (a *TestNodeDb) StoreContextFileTransfer(c *gripdata.ContextFileTransfer) (*gripdata.ContextFileTransferWrap, error) {
	return nil, nil
}
func (a *TestNodeDb) DeleteContextFile(c *gripdata.ContextFileWrap) error {
	return nil
}
func (a *TestNodeDb) GetContextHeads(cid []byte) []*gripdata.ContextFileWrap {
	return nil
}
func (a *TestNodeDb) GetContextLeaves(cid []byte, covered bool, index bool) []*gripdata.ContextFileWrap {
	return nil
}
func (a *TestNodeDb) GetCoveredSnapshots(cid []byte) []*gripdata.ContextFileWrap {
	return nil
}
func (t *TestNodeDb) CheckUpdateStorageUsed(a *gripdata.Account, fsize uint64) error {
	return nil
}
func (t *TestNodeDb) FreeStorageUsed(a *gripdata.Account, fsize uint64) (*gripdata.Account, error) {
	return nil, nil
}
func (t *TestNodeDb) SetAccountMessage(a *gripdata.Account, msg string) error {
	return nil
}
func (t *TestNodeDb) StoreAccount(a *gripdata.Account) error {
	return nil
}
func (t *TestNodeDb) GetAccount(s string) *gripdata.Account {
	return nil
}
func (t *TestNodeDb) StoreNodeAccountKey(ak *gripdata.NodeAccountKey) error {
	return nil
}
func (t *TestNodeDb) GetNodeAccountKey(key string) *gripdata.NodeAccountKey {
	return nil
}
func (t *TestNodeDb) StoreNodeAccount(na *gripdata.NodeAccount) error {
	return nil
}
func (t *TestNodeDb) GetNodeAccount(id []byte) *gripdata.NodeAccount {
	return nil
}
func (t *TestNodeDb) IncrNumberContexts(a *gripdata.Account) error {
	return nil
}
func (t *TestNodeDb) IncrNumberNodes(a *gripdata.Account) error {
	return nil
}
func (t *TestNodeDb) GetContextFileDeleted(dig []byte) *gripdata.DeletedContextFile {
	return nil
}
func (t *TestNodeDb) GetFileTransfersForNode(id []byte, max int) []*gripdata.ContextFileTransferWrap {
	return nil
}
func (t *TestNodeDb) DeleteContextFileTransfer(nodeid []byte, confiledig []byte) (string, error) {
	return "", nil
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
	_, err2 := grip.VerifyNode(&n, &db)
	if err2 != nil {
		t.Error(err2)
	}
	fmt.Println(hex.EncodeToString(n.ID))
	fmt.Println(hex.EncodeToString(pr.ID))
}
