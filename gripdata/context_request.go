package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//ContextRequest is used to request a node participate
//within a context.
type ContextRequest struct {
	ContextDig         []byte //The Dig of the Context record
	NodeID             []byte //Node ID of creating node
	TargetNodeID       []byte //The id of the node the request is for
	FullRepo           bool   //Should TargetNodeID keep a full snapshot of the context
	AllowContextSource bool   //TargetNodeID can source new data for the context
	AllowContextNode   bool   //TargetNodeID can add new nodes to the context
	AllowNewLogin      bool   //TargetNodeID can create logins for the context
	CacheMode          uint32 //Indicates how data is cached for non-full repo
	MetaData           string //Generic metadata string
	Dig                []byte //This record's digest
	Sig                []byte //Signed by NodeID private key
}

//Digest ContextRequest
func (a *ContextRequest) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashBytes(h, a.NodeID)
	gripcrypto.HashBool(h, a.AllowContextNode)
	gripcrypto.HashBool(h, a.AllowContextSource)
	gripcrypto.HashBool(h, a.AllowNewLogin)
	gripcrypto.HashUint32(h, a.CacheMode)
	gripcrypto.HashBytes(h, a.ContextDig)
	gripcrypto.HashBool(h, a.FullRepo)
	gripcrypto.HashString(h, a.MetaData)
	gripcrypto.HashBytes(h, a.TargetNodeID)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig ContextRequest
func (a *ContextRequest) SetSig(b []byte) {
	a.Sig = b
}

//GetSig ContextRequest
func (a *ContextRequest) GetSig() []byte {
	return a.Sig
}

//GetNodeID ContextRequest
func (a *ContextRequest) GetNodeID() []byte {
	return a.NodeID
}

//GetDig ContextRequest
func (a *ContextRequest) GetDig() []byte {
	return a.Dig
}

//SetNodeID ContextRequest
func (a *ContextRequest) SetNodeID(id []byte) {
	a.NodeID = id
}

//checkContextRequest just forces interface check
func checkContextRequest() {
	var f ContextRequest
	gripcrypto.Sign(&f, nil)
}
