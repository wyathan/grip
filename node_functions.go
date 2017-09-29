package grip

import (
	"crypto/sha512"
	"errors"
	"reflect"
	"time"

	"encoding/base64"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

func ListShareWithMe() []gripdata.ShareNodeInfo {
	return nil
}

//NewShareNode states that this node's Node data should be
//shared with another node
func NewShareNode(n *gripdata.ShareNodeInfo, db NodeNetdb) error {
	id := n.NodeID
	if id == nil {
		return errors.New("NodeID is nil")
	}
	err := SignNodeSig(n, db)
	if err != nil {
		return err
	}
	err = db.StoreShareNodeInfo(n)
	if err != nil {
		return err
	}
	nd := db.GetNode(id)
	if nd == nil {
		return errors.New("This node is not found")
	}
	err = SendAllToShareWith(nd, db)
	if err != nil {
		return err
	}
	err = SendAllToShareWith(n, db)
	if err != nil {
		return err
	}
	err = SendAllSharesToNew(id, n.TargetNodeID, db)
	if err != nil {
		return err
	}
	return nil
}

//NewUseShareNodeKey states that this node's Node data should be
//shared with another node
func NewUseShareNodeKey(n *gripdata.UseShareNodeKey, db NodeNetdb) error {
	if n.Key == "" {
		return errors.New("Key cannot be empty")
	}
	id := n.NodeID
	if id == nil {
		return errors.New("NodeID is nil")
	}
	err := SignNodeSig(n, db)
	if err != nil {
		return err
	}
	err = db.StoreUseShareNodeKey(n)
	if err != nil {
		return err
	}
	nd := db.GetNode(id)
	if nd == nil {
		return errors.New("This node is not found")
	}
	err = SendAllToAllKey(nd, n.Key, db)
	if err != nil {
		return err
	}
	err = SendAllToAllKey(n, n.Key, db)
	if err != nil {
		return err
	}
	return nil
}

//SendAllToAllKey sends new data to all nodes on share list
func SendAllToAllKey(s gripcrypto.SignInf, k string, db NodeNetdb) error {
	shl := FindAllToShareWithFromKey(k, db)
	for _, sh := range shl {
		err := CreateNewSend(s, sh, db)
		if err != nil {
			return err
		}
	}
	return nil
}

//FindAllToShareWithFromKey get all nodes to send data to based on key
func FindAllToShareWithFromKey(k string, db NodeNetdb) [][]byte {
	var s [][]byte
	kl := db.ListShareNodeKey(k)
	for _, sn := range kl {
		s = append(s, FindAllToShareWith(sn.NodeID, db)...)
	}
	return s
}

//FindAllToShareWith get all the nodes to share updates about a node
func FindAllToShareWith(id []byte, db NodeNetdb) [][]byte {
	sl := db.ListShareNodeInfo(id)
	km := make(map[string]bool)
	bm := make(map[string][]byte)
	for _, sr := range sl {
		bm[base64.StdEncoding.EncodeToString(sr.TargetNodeID)] = sr.TargetNodeID
		if sr.Key != "" {
			km[sr.Key] = true
		}
	}
	for k := range km {
		//Find all nodes that used key
		skl := db.ListUseShareNodeKey(k)
		for _, sk := range skl {
			bm[base64.StdEncoding.EncodeToString(sk.NodeID)] = sk.NodeID
		}
	}
	var s [][]byte
	for _, v := range bm {
		s = append(s, v)
	}
	return s
}

//SendAllSharesToNew sends all data to a new share node
func SendAllSharesToNew(from []byte, to []byte, db NodeNetdb) error {
	shl := db.ListShareNodeInfo(from)
	km := make(map[string]bool)
	for _, sr := range shl {
		nd := db.GetNode(sr.TargetNodeID)
		err := CreateNewSend(nd, to, db)
		if err != nil {
			return err
		}
		err = CreateNewSend(&sr, to, db)
		if err != nil {
			return err
		}
		if sr.Key != "" {
			km[sr.Key] = true
		}
	}
	for k := range km {
		//Find all nodes that used key
		skl := db.ListUseShareNodeKey(k)
		for _, sk := range skl {
			nd := db.GetNode(sk.NodeID)
			err := CreateNewSend(nd, to, db)
			if err != nil {
				return err
			}
			err = CreateNewSend(&sk, to, db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//SendAllToShareWith sends new data to all nodes on share list
func SendAllToShareWith(s gripcrypto.SignInf, db NodeNetdb) error {
	shl := FindAllToShareWith(s.GetNodeID(), db)
	for _, sh := range shl {
		err := CreateNewSend(s, sh, db)
		if err != nil {
			return err
		}
	}
	return nil
}

//CreateNewSend indicates this data should be sent to this node
func CreateNewSend(v gripcrypto.SignInf, sendto []byte, db Netdb) error {
	var s gripdata.SendData
	s.Dig = v.GetDig()
	s.TargetID = sendto
	s.TypeName = reflect.TypeOf(v).String()
	s.Timestamp = uint64(time.Now().UnixNano())
	err := db.StoreSendData(&s)
	if err != nil {
		return err
	}
	return nil
}

//CreateNewNode creates a new key pair for the node, saves
//the private node data, generates the id, and signs it
func CreateNewNode(pr *gripdata.MyNodePrivateData, n *gripdata.Node, db Nodedb) error {
	if pr.PrivateKey == nil || n.PublicKey == nil {
		prv, pub, err := gripcrypto.GenerateECDSAKeyPair()
		if err != nil {
			return err
		}
		pr.PrivateKey = prv
		n.PublicKey = pub
		h := sha512.New()
		gripcrypto.HashBytes(h, n.PublicKey)
		n.ID = h.Sum(nil)
		pr.ID = n.ID
	}
	err := gripcrypto.Sign(n, pr.PrivateKey)
	if err != nil {
		return err
	}
	err = db.StoreMyPrivateNodeData(n, pr)
	if err != nil {
		return err
	}
	return nil
}

//SignNodeSig sign a new sign interface
func SignNodeSig(s gripcrypto.SignInf, db Nodedb) error {
	_, privdata := db.GetPrivateNodeData()
	s.SetNodeID(privdata.ID)
	err := gripcrypto.Sign(s, privdata.PrivateKey)
	if err != nil {
		return err
	}
	return nil
}

//VerifyNodeSig verify a node signed a sign interface
func VerifyNodeSig(s gripcrypto.SignInf, db Nodedb) error {
	nodeid := s.GetNodeID()
	if nodeid == nil {
		return errors.New("NodeID was not set")
	}
	nodedata := db.GetNode(nodeid)
	if nodedata == nil {
		return errors.New("Node was not found in db")
	}
	if !gripcrypto.Verify(s, nodedata.PublicKey) {
		return errors.New("Signature was not valid")
	}
	return nil
}
