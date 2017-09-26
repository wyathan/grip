package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//ContextFile file data pertaining to a context
type ContextFile struct {
	Index       bool
	Snapshot    bool
	DependsOn   [][]byte
	Path        string
	ContextUser string //If a logged in context user submitted this
	NodeID      []byte //Node ID of creating node
	CreatedOn   uint64 //Time stamp
	Context     []byte //Digest of the context this file bellongs to
	Dig         []byte //This record's digest
	Sig         []byte //Signed by NodeID private key
}

//Digest ContextFile
func (a *ContextFile) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashBool(h, a.Index)
	gripcrypto.HashBool(h, a.Snapshot)
	gripcrypto.HashString(h, a.ContextUser)
	gripcrypto.HashBytes(h, a.NodeID)
	gripcrypto.HashUint64(h, a.CreatedOn)
	gripcrypto.HashBytes(h, a.Context)
	gripcrypto.HashFile(h, a.Path)
	if a.DependsOn == nil {
		gripcrypto.HashUint32(h, 0)
	} else {
		gripcrypto.HashUint32(h, uint32(len(a.DependsOn)))
		for _, d := range a.DependsOn {
			gripcrypto.HashBytes(h, d)
		}
	}
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig ContextFile
func (a *ContextFile) SetSig(b []byte) {
	a.Sig = b
}

//GetSig ContextFile
func (a *ContextFile) GetSig() []byte {
	return a.Sig
}

//GetNodeID ContextFile
func (a *ContextFile) GetNodeID() []byte {
	return a.NodeID
}

//GetDig ContextFile
func (a *ContextFile) GetDig() []byte {
	return a.Dig
}

//SetNodeID ContextFile
func (a *ContextFile) SetNodeID(id []byte) {
	a.NodeID = id
}

//checkContextFunctionsImplSign just forces interface check
func checkContextFunctionsImplSign() {
	var f ContextFile
	gripcrypto.Sign(&f, nil)
}
