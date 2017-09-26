package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//Context created by a node.
type Context struct {
	Name      string //The name of the context
	NodeID    []byte //The id of the node that created it
	CreatedOn uint64 //Timestamp when created
	Dig       []byte //This record's digest
	Sig       []byte //Signed by NodeID private key
}

//Digest Context
func (a *Context) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashString(h, a.Name)
	gripcrypto.HashBytes(h, a.NodeID)
	gripcrypto.HashUint64(h, a.CreatedOn)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig Context
func (a *Context) SetSig(b []byte) {
	a.Sig = b
}

//GetSig Context
func (a *Context) GetSig() []byte {
	return a.Sig
}

//GetNodeID Context
func (a *Context) GetNodeID() []byte {
	return a.NodeID
}

//GetDig Context
func (a *Context) GetDig() []byte {
	return a.Dig
}

//SetNodeID Context
func (a *Context) SetNodeID(id []byte) {
	a.NodeID = id
}

//checkContext just forces interface check
func checkContext() {
	var f Context
	gripcrypto.Sign(&f, nil)
}
