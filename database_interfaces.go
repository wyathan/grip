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
}

//Nodedb used for storing/getting node data from db
type Nodedb interface {
	GetNode(id []byte) *gripdata.Node
	GetPrivateNodeData() (*gripdata.Node, *gripdata.MyNodePrivateData)
	GetShareNodeInfo(id []byte, target []byte) *gripdata.ShareNodeInfo
	StoreNode(n *gripdata.Node) error
	StoreMyPrivateNodeData(n *gripdata.Node, pr *gripdata.MyNodePrivateData) error
	StoreShareNodeInfo(sn *gripdata.ShareNodeInfo) error
	ListNodes() []gripdata.Node                           //List all known nodes
	ListShareNodeInfo(id []byte) []gripdata.ShareNodeInfo //list all ShareNodeInfo from node
	ListShareNodeKey(s string) []gripdata.ShareNodeInfo   //List all ShareNodeInfo with key
	StoreUseShareNodeKey(k *gripdata.UseShareNodeKey) error
	ListUseShareNodeKey(k string) []gripdata.UseShareNodeKey //List all that have used key
}

//Netdb used for network db access
type Netdb interface {
	StoreSendData(s *gripdata.SendData) error
	GetSendData(target []byte, max int) []gripdata.SendData //get all send data for target node
	DeleteSendData(s *gripdata.SendData) error              //Data has been setnt to the node
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
