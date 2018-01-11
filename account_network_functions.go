package grip

import (
	"time"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
	"github.com/wyathan/grip/griperrors"
)

type ANAKProc func(s *gripdata.AssociateNodeAccountKey, db DB) (bool, error)

func AssKeyProc(np ANAKProc) SProc {
	return func(s gripcrypto.SignInf, db DB) (bool, error) {
		switch n := s.(type) {
		case *gripdata.AssociateNodeAccountKey:
			return np(n, db)
		}
		return false, griperrors.WrongType
	}
}

func CheckNodeAccontKey(s *gripdata.AssociateNodeAccountKey, db DB) (bool, error) {
	ak := db.GetNodeAccountKey(s.Key)
	if ak == nil {
		return false, griperrors.NodeAccountKeyNotFound
	}
	if ak.Used {
		return false, griperrors.NodeAccountKeyUsed
	}
	nw := uint64(time.Now().UnixNano())
	if nw > ak.Expiration {
		return false, griperrors.NodeAccountKeyExpired
	}
	return true, nil
}

func CheckAccountAllowsKeys(s *gripdata.AssociateNodeAccountKey, db DB) (bool, error) {
	ak := db.GetNodeAccountKey(s.Key)
	a := db.GetAccount(ak.AccountID)
	if a == nil {
		return false, griperrors.AccountNotFound
	}
	if !a.Enabled {
		return false, griperrors.AccountNotEnabled
	}
	if !a.AllowNodeAcocuntKey {
		return false, griperrors.AccountKeyNotAllowed
	}
	if a.NumberNodes >= a.MaxNodes {
		return false, griperrors.MaxNodesForAccount
	}
	return true, nil
}

func IncrNumberNodes(s *gripdata.AssociateNodeAccountKey, db DB) (bool, error) {
	ak := db.GetNodeAccountKey(s.Key)
	a := db.GetAccount(ak.AccountID)
	return true, db.IncrNumberNodes(a)
}

func CreateNewNodeAccount(s *gripdata.AssociateNodeAccountKey, db DB) (bool, error) {
	ak := db.GetNodeAccountKey(s.Key)
	var na gripdata.NodeAccount
	na.AccountID = ak.AccountID
	na.Enabled = true
	na.NodeID = s.NodeID
	return true, db.StoreNodeAccount(&na)
}

//IncomingNodeAccountKey process incoming ShareNodeInfoKeys
func IncomingNodeAccountKey(s *gripdata.AssociateNodeAccountKey, db DB) error {
	var v SProcChain
	v.Push(VerifyNodeSig)
	v.Push(AssKeyProc(CheckNodeAccontKey))
	v.Push(AssKeyProc(CheckAccountAllowsKeys))
	v.Push(AssKeyProc(IncrNumberNodes))
	v.Push(AssKeyProc(CreateNewNodeAccount))
	_, err := v.F(s, db)
	return err
}
