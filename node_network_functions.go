package grip

import (
	"bytes"
	"crypto/sha512"
	"errors"
	"log"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
	"github.com/wyathan/grip/griperrors"
)

type NProc func(c *gripdata.Node, db DB) (bool, error)

func NodeProc(np NProc) SProc {
	return func(s gripcrypto.SignInf, db DB) (bool, error) {
		switch n := s.(type) {
		case *gripdata.Node:
			return np(n, db)
		}
		return false, griperrors.WrongType
	}
}

func StoreNode(c *gripdata.Node, db DB) (bool, error) {
	return true, db.StoreNode(c)
}

func CreateNodeEphemera(n *gripdata.Node, db DB) (bool, error) {
	return true, db.CreateNodeEphemera(n.ID, n.Connectable)
}

func LocallyCreated(n gripcrypto.SignInf, db DB) (bool, error) {
	_, pr := db.GetPrivateNodeData()
	if bytes.Equal(pr.ID, n.GetNodeID()) {
		return false, nil
	}
	return true, nil
}

//IncomingNode process incoming node data
//Forward to all nodes that it has shared with
func IncomingNode(n *gripdata.Node, db DB) error {
	var p SProcChain
	p.Push(LocallyCreated)
	p.Push(NodeProc(VerifyNode))
	p.Push(NodeProc(StoreNode))
	p.Push(NodeProc(CreateNodeEphemera))
	p.Push(SendAllToShareWithMe)

	_, err := p.F(n, db)
	return err
}

func TargetNil(s *gripdata.ShareNodeInfo, db DB) (bool, error) {
	tn := s.TargetNodeID
	if tn == nil {
		return false, griperrors.TargetNodeNil
	}
	return true, nil
}

func CreateShareNodeAccount(s *gripdata.ShareNodeInfo, db DB) (bool, error) {
	_, pr := db.GetPrivateNodeData()
	createAutoAccount(pr, s.NodeID, db)
	if !IsAccountEnabled(s, db) {
		return false, griperrors.AccountNotEnabled
	}
	return true, nil
}

func ShareWithMe(s *gripdata.ShareNodeInfo, db DB) (bool, error) {
	_, pr := db.GetPrivateNodeData()
	if bytes.Equal(pr.ID, s.TargetNodeID) {
		log.Println("Sharing node data with me")
		return false, nil
	}
	return true, nil
}

//SendAllToShareWith sends new data to all nodes on share list
func SendAllToShareWith(s gripcrypto.SignInf, sourceid []byte, db DB) error {
	shl := FindAllToShareWith(sourceid, db)
	for _, sh := range shl {
		err := CreateNewSend(s, sh, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func SendTargetToShareWith(s *gripdata.ShareNodeInfo, db DB) (bool, error) {
	gn := db.GetNode(s.TargetNodeID)
	return true, SendAllToShareWith(gn, s.NodeID, db)
}

//IncomingShareNode process incoming ShareNodeInfo
func IncomingShareNode(s *gripdata.ShareNodeInfo, db DB) error {
	var p SProcChain
	n := db.GetNode(s.NodeID)
	p.Push(LocallyCreated)
	p.Push(ShareNodeProc(TargetNil))
	p.Push(VerifyNodeSig)
	p.Push(ShareNodeProc(CreateShareNodeAccount))
	p.Push(ShareNodeProc(StoreShareNodeInfo))
	p.Push(ShareNodeProc(ShareWithMe))
	p.PushV(SendAllToShareWithMe, n)
	p.Push(ShareNodeProc(SendTargetToShareWith))
	p.Push(SendAllToShareWithMe)
	p.Push(ShareNodeProc(SendAllShareNodeInfo))

	_, err := p.F(s, db)
	return err
}

func CreateUseKeyAccount(s *gripdata.UseShareNodeKey, db DB) (bool, error) {
	_, pr := db.GetPrivateNodeData()
	createAutoAccount(pr, s.NodeID, db)
	if !IsAccountEnabled(s, db) {
		return false, griperrors.AccountNotEnabled
	}
	return true, nil
}

//SendAllToAllKey sends new data to all nodes on share list
func SendAllToAllKey(s gripcrypto.SignInf, k string, db DB) error {
	shl := FindAllToShareWithFromKey(k, db)
	for _, sh := range shl {
		err := CreateNewSend(s, sh, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func PrepSendAllToAllKey(k string) SProc {
	return func(s gripcrypto.SignInf, db DB) (bool, error) {
		return true, SendAllToAllKey(s, k, db)
	}
}

func SendAllSharesToNewKeyNode(k *gripdata.UseShareNodeKey, db DB) (bool, error) {
	kl := db.ListShareNodeKey(k.Key)
	for _, v := range kl {
		err := SendAllSharesToNew(v.NodeID, k.NodeID, db)
		if err != nil {
			return true, err
		}
	}
	return true, nil
}

//IncomingUseShareNodeKey processes an incoming UseShareNodeKey
func IncomingUseShareNodeKey(k *gripdata.UseShareNodeKey, db DB) error {
	if k.Key == "" {
		return griperrors.ShareNodeKeyEmpty
	}
	n := db.GetNode(k.NodeID)
	var p SProcChain
	p.Push(LocallyCreated)
	p.Push(VerifyNodeSig)
	p.Push(UseKeyProc(CreateUseKeyAccount))
	p.Push(UseKeyProc(StoreUseShareNodeKey))
	p.PushV(PrepSendAllToAllKey(k.Key), n)
	p.Push(PrepSendAllToAllKey(k.Key))
	p.Push(UseKeyProc(SendAllSharesToNewKeyNode))

	_, err := p.F(k, db)
	return err
}

//VerifyNode verify new node data
func VerifyNode(n *gripdata.Node, db DB) (bool, error) {
	if n.PublicKey == nil || n.ID == nil {
		return false, errors.New("Public key and ID must be specified")
	}
	h := sha512.New()
	gripcrypto.HashBytes(h, n.PublicKey)
	tid := h.Sum(nil)
	if !bytes.Equal(n.ID, tid) {
		return false, errors.New("ID does not match public key")
	}
	if !gripcrypto.Verify(n, n.PublicKey) {
		return false, errors.New("Failed to verify signature")
	}
	return true, nil
}

//GetNodeAccount see if this node id has an account
func GetNodeAccount(id []byte, db Accountdb) *gripdata.Account {
	na := db.GetNodeAccount(id)
	if na == nil {
		return nil
	}
	if !na.Enabled {
		return nil
	}
	return db.GetAccount(na.AccountID)
}

//IsIDAccountEnabled see if this node id has account enabled
func IsIDAccountEnabled(id []byte, db Accountdb) bool {
	a := GetNodeAccount(id, db)
	if a == nil {
		return false
	}
	return a.Enabled
}

//IsAccountEnabled see if this signInf has account enabled
func IsAccountEnabled(s gripcrypto.SignInf, db Accountdb) bool {
	return IsIDAccountEnabled(s.GetNodeID(), db)
}
