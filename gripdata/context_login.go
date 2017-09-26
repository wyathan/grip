package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//ContextLogin can be used by nodes that have permission to source context
//data.  If a user logs in with the id the node can source new data for a
//context.
type ContextLogin struct {
	ContextDig []byte //The Dig of the ContextReqest
	Login      string //The login user name
	PWDig      []byte //The password hash/digest
	Expiration uint64 //The time this key expires
	NodeID     []byte //ID of the node that created the login
	Dig        []byte //This record's digest
	Sig        []byte //Signed by NodeID private key
}

//Digest ContextLogin
func (a *ContextLogin) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashBytes(h, a.ContextDig)
	gripcrypto.HashUint64(h, a.Expiration)
	gripcrypto.HashString(h, a.Login)
	gripcrypto.HashBytes(h, a.NodeID)
	gripcrypto.HashBytes(h, a.PWDig)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig ContextLogin
func (a *ContextLogin) SetSig(b []byte) {
	a.Sig = b
}

//GetSig ContextLogin
func (a *ContextLogin) GetSig() []byte {
	return a.Sig
}

//GetNodeID ContextLogin
func (a *ContextLogin) GetNodeID() []byte {
	return a.NodeID
}

//GetDig ContextLogin
func (a *ContextLogin) GetDig() []byte {
	return a.Dig
}

//SetNodeID ContextLogin
func (a *ContextLogin) SetNodeID(id []byte) {
	a.NodeID = id
}

//checkContextLogin just forces interface check
func checkContextLogin() {
	var f ContextLogin
	gripcrypto.Sign(&f, nil)
}
