package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//ContextResponse the response of a the TargetNodeID node to a ContextRequest
//The permissions are a result of AND'ing the request with the response.
type ContextResponse struct {
	ContextDig    []byte //The Dig of the Context
	TargetNodeID  []byte //The id of the target node for the request (this signing node)
	RequestDig    []byte //The Dig of the ContextRequest repsonding to
	FullRepo      bool   //Should TargetNodeID keep a full snapshot of the context
	ContextSource bool   //TargetNodeID can source new data for the context
	ContextNode   bool   //TargetNodeID can add new nodes to the context
	NewLogin      bool   //TargetNodeID can create logins for the context
	CacheMode     uint32 //Indicates how data is cached for non-full repo
	MetaData      string //Generic metadata string
	Dig           []byte //This record's digest
	Sig           []byte //Signed by request's TargetNodeIDdeID private key
}

//Digest ContextResponse
func (a *ContextResponse) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashUint32(h, a.CacheMode)
	gripcrypto.HashBytes(h, a.ContextDig)
	gripcrypto.HashBool(h, a.ContextNode)
	gripcrypto.HashBool(h, a.ContextSource)
	gripcrypto.HashBool(h, a.FullRepo)
	gripcrypto.HashString(h, a.MetaData)
	gripcrypto.HashBool(h, a.NewLogin)
	gripcrypto.HashBytes(h, a.RequestDig)
	gripcrypto.HashBytes(h, a.TargetNodeID)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig ContextResponse
func (a *ContextResponse) SetSig(b []byte) {
	a.Sig = b
}

//GetSig ContextResponse
func (a *ContextResponse) GetSig() []byte {
	return a.Sig
}

//GetNodeID ContextResponse
func (a *ContextResponse) GetNodeID() []byte {
	return a.TargetNodeID
}

//GetDig ContextResponse
func (a *ContextResponse) GetDig() []byte {
	return a.Dig
}

//SetNodeID ContextResponse
func (a *ContextResponse) SetNodeID(id []byte) {
	a.TargetNodeID = id
}

//checkContextResponse just forces interface check
func checkContextResponse() {
	var f ContextResponse
	gripcrypto.Sign(&f, nil)
}
