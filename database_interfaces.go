package grip

import "github.com/wyathan/grip/gripdata"

//Accountdb used for storing/getting account data from db
type Accountdb interface {
	StoreAccount(a *gripdata.Account) error
	GetAccount(s string) *gripdata.Account
	StoreNodeAccountKey(ak *gripdata.NodeAccountKey) error
	GetNodeAccountKey(key string) *gripdata.NodeAccountKey
	StoreNodeAccount(na *gripdata.NodeAccount) error
	GetNodeAccount(id []byte) *gripdata.NodeAccount
	IncrNumberContexts(a *gripdata.Account) error
	IncrNumberNodes(a *gripdata.Account) error
	CheckUpdateStorageUsed(a *gripdata.Account, fsize uint64) error
	FreeStorageUsed(a *gripdata.Account, fsize uint64) (*gripdata.Account, error)
	SetAccountMessage(a *gripdata.Account, msg string) error
}

//Nodedb used for storing/getting node data from db
type Nodedb interface {
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
}

//Netdb used for network db access
type Netdb interface {
	StoreSendData(s *gripdata.SendData) error
	StoreRejectedSendData(s *gripdata.RejectedSendData) error
	//get all send data for target node, must be sorted by Timestamp
	GetSendData(target []byte, max int) []gripdata.SendData
	//Data has been setnt to the node, no error if missing
	//return bool as true if successfully deleted, false if it didn't exist
	DeleteSendData(d []byte, to []byte) (bool, error)
	GetDigestData(d []byte) interface{}
	//NodeEphemera.Connectable == true
	//NodeEphemera.Connected == false
	//NodeEphemera.NextAttempt <= curtime
	//exists some SendData.TargetID == NodeEphemera.ID
	GetConnectableNodesWithSendData(max int, curtime uint64) []gripdata.NodeEphemera
	//Get nodes that are connectable that we have sent a ShareNodeInfo with a Key,
	//So we can get new UseShareNodeKeys that may have been submitted to that node.
	GetConnectableNodesWithShareNodeKey(max int, curtime uint64) []gripdata.NodeEphemera
	//Get nodes that we have sent a UseShareNodeKey so we can connect to them
	//if they are connectable
	GetConnectableUseShareKeyNodes(max int, curtime uint64) []gripdata.NodeEphemera
	//Get nodes with file transfers to them.
	GetConnectableWithFileTransfers(max int, curtime uint64) []gripdata.NodeEphemera
	GetConnectableAny(max int, curtime uint64) []gripdata.NodeEphemera
	GetAllConnected() []gripdata.NodeEphemera
	CreateNodeEphemera(id []byte, connectable bool) error
	SetNodeEphemeraNextConnection(id []byte, last uint64, next uint64) error
	ClearAllConnected()
	CanNodeEphemeraGoPending(id []byte) bool
	SetNodeEphemeraConnected(incomming bool, id []byte, curtime uint64) error
	SetNodeEphemeraClosed(id []byte) error
	//Get all the ContextRequest/ContextResponse pairs for this node
	GetNodeContextPairs(id []byte) map[string]*gripdata.ContextPairWrap
	//See if there are ContextFileTransfers for this context
	//For currently connected nodes.  If there are, request
	GetFileTransfersForConnected(conid []byte) []*gripdata.ContextFileTransferWrap
}

//Contextdb sotre/load context data
type Contextdb interface {
	StoreContext(c *gripdata.Context) error
	GetContext(id []byte) *gripdata.Context
	GetContextRequests(id []byte) []*gripdata.ContextRequest
	StoreContextRequest(c *gripdata.ContextRequest) error
	GetContextRequest(cid []byte, tgtid []byte) *gripdata.ContextRequest
	StoreContextResponse(c *gripdata.ContextResponse) error
	GetContextResponse(cid []byte, tgtid []byte) *gripdata.ContextResponse
	GetContextResponses(cid []byte) []*gripdata.ContextResponse
	GetContextFileByDepDataDig(d []byte) *gripdata.ContextFileWrap
	//Never ever be able to access these as valid data!  Only for debug!
	StoreVeryBadContextFile(c *gripdata.ContextFile) error
	StoreContextFile(c *gripdata.ContextFile) (*gripdata.ContextFileWrap, error)
	GetAllThatDependOn(cid []byte, dig []byte) []*gripdata.ContextFileWrap
	StoreContextFileTransfer(c *gripdata.ContextFileTransfer) (*gripdata.ContextFileTransferWrap, error)
	//Do not return error if it doesn't exist, string is the path to the file to delete, it should
	//be nil if the file should not be deleted yet.
	DeleteContextFileTransfer(nodeid []byte, confiledig []byte) (string, error)
	GetFileTransfersForNode(id []byte, max int) []*gripdata.ContextFileTransferWrap
	SetContextFileForTransfer(cf *gripdata.ContextFile) error
	//Make sure it creates a DeletedContextFile record
	DeleteContextFile(c *gripdata.ContextFileWrap) error
	//Gets a deleted record based on the ContextFileDig
	GetContextFileDeleted(dig []byte) *gripdata.DeletedContextFile
	GetContextHeads(cid []byte) []*gripdata.ContextFileWrap
	GetContextLeaves(cid []byte, covered bool, index bool) []*gripdata.ContextFileWrap
	//Sort by depth and size
	GetCoveredSnapshots(cid []byte) []*gripdata.ContextFileWrap
}

//NodeContextdb implements both Nodedb and Contextdb
type NodeContextdb interface {
	Nodedb
	Contextdb
}

//NodeNetContextdb implements node, net, and context dbs
type NodeNetContextdb interface {
	Nodedb
	Netdb
	Contextdb
}

//NodeAccountdb implements both Nodedb and Accountdb
type NodeAccountdb interface {
	Nodedb
	Accountdb
}

//NodeNetdb implements both Nodedb and Netdb
type NodeNetdb interface {
	Nodedb
	Netdb
}

//NodeNetAccountdb implements Nodedb, Netdb, and Accountdb
type NodeNetAccountdb interface {
	Nodedb
	Netdb
	Accountdb
}

//DB implements all database interfaces
type DB interface {
	Nodedb
	Netdb
	Contextdb
	Accountdb
}
