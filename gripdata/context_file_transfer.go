package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//ContextFileTransfer Indicates that a context file should be
//transfered to a node.  These are created by nodes participating
//in a context to make sure new nodes get the context files they need
//ContextFile file data pertaining to a context
type ContextFileTransfer struct {
	Context    []byte //The context the file is from
	DataDepDig []byte //The DataDepDig for the file to transfer
	TrasnferTo []byte //The id of the node to transfer the file to
	Dig        []byte //Just digest of Context, DataDebDig, and TransferTo
	//this way we only keep one transfer per file and not
	//duplicates for every node that has a full repo

	NodeID []byte //Node ID of creating node, just needed to validate sig
	Sig    []byte //Signed by NodeID private key
}

//Digest ContextFileTransfer
func (a *ContextFileTransfer) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashBytes(h, a.Context)
	gripcrypto.HashBytes(h, a.DataDepDig)
	gripcrypto.HashBytes(h, a.TrasnferTo)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig ContextFileTransfer
func (a *ContextFileTransfer) SetSig(b []byte) {
	a.Sig = b
}

//GetSig ContextFileTransfer
func (a *ContextFileTransfer) GetSig() []byte {
	return a.Sig
}

//GetNodeID ContextFileTransfer
func (a *ContextFileTransfer) GetNodeID() []byte {
	return a.NodeID
}

//GetDig ContextFileTransfer
func (a *ContextFileTransfer) GetDig() []byte {
	return a.Dig
}

//SetNodeID ContextFileTransfer
func (a *ContextFileTransfer) SetNodeID(id []byte) {
	a.NodeID = id
}

//checkContextFileTransfer just forces interface check
func checkContextFileTransfer() {
	var f ContextFileTransfer
	gripcrypto.Sign(&f, nil)
}
