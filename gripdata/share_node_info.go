package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//ShareNodeInfo says that this node is ok with this other node
//knowing about it
type ShareNodeInfo struct {
	TargetNodeID []byte //The node that can know about NodeID
	NodeID       []byte //The id of the node data to share
	MetaData     string //Generic metadata string
	Key          string
	Dig          []byte //This record's digest
	Sig          []byte //Signed by NodeID private key
}

//Digest ShareNodeInfo
func (a *ShareNodeInfo) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashString(h, a.MetaData)
	gripcrypto.HashBytes(h, a.NodeID)
	gripcrypto.HashBytes(h, a.TargetNodeID)
	gripcrypto.HashString(h, a.Key)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig ShareNodeInfo
func (a *ShareNodeInfo) SetSig(b []byte) {
	a.Sig = b
}

//GetSig ShareNodeInfo
func (a *ShareNodeInfo) GetSig() []byte {
	return a.Sig
}

//GetNodeID ShareNodeInfo
func (a *ShareNodeInfo) GetNodeID() []byte {
	return a.NodeID
}

//GetDig ShareNodeInfo
func (a *ShareNodeInfo) GetDig() []byte {
	return a.Dig
}

//SetNodeID ShareNodeInfo
func (a *ShareNodeInfo) SetNodeID(id []byte) {
	a.NodeID = id
}

//checkShareNodeInfo just forces interface check
func checkShareNodeInfo() {
	var f ShareNodeInfo
	gripcrypto.Sign(&f, nil)
}
