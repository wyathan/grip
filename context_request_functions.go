package grip

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

//CheckNonEscalating Make sure the requesting node is not requesting more
//permissions for the target node than it has itself
func CheckNonEscalating(src *gripdata.ContextRequest, tgt *gripdata.ContextRequest) bool {
	if !src.AllowContextNode {
		return false
	}
	if !src.AllowContextSource && tgt.AllowContextSource {
		return false
	}
	if !src.AllowNewLogin && tgt.AllowNewLogin {
		return false
	}
	return true
}

func canNonOwnerCreateContextRequest(myid []byte, c *gripdata.ContextRequest, ctx *gripdata.Context, db DB) error {
	//See if we have permission to add others
	myreq := db.GetContextRequest(ctx.Dig, myid)
	myrsp := db.GetContextResponse(ctx.Dig, myid)
	if !(myreq.AllowContextNode && myrsp.ContextNode) {
		return errors.New("Not allowed to request new nodes for context")
	}
	if !CheckNonEscalating(myreq, c) {
		return errors.New("You cannot escilate permissions for a new node")
	}
	return nil
}

func canCreateContextRequest(c *gripdata.ContextRequest, db DB) (*gripdata.Context, error) {
	myn, _ := db.GetPrivateNodeData()
	//Owner can do what he likes
	ctx := db.GetContext(c.ContextDig)
	if ctx == nil {
		return nil, errors.New("Context is not found")
	}
	if bytes.Equal(ctx.NodeID, myn.ID) {
		//I'm the owner, it's ok
		return ctx, nil
	}
	err := canNonOwnerCreateContextRequest(myn.ID, c, ctx, db)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func validateNewContextRequest(c *gripdata.ContextRequest, db DB) (*gripdata.Context, error) {
	tid := db.GetNode(c.TargetNodeID)
	if tid == nil {
		return nil, errors.New("Unknown target node")
	}
	ctx, err := canCreateContextRequest(c, db)
	if err != nil {
		return nil, err
	}
	rqst := db.GetContextRequest(c.ContextDig, c.TargetNodeID)
	if rqst != nil {
		return nil, errors.New("There is already a ContextRequest for this node")
	}
	return ctx, nil
}

func signAndStoreContextRequest(c *gripdata.ContextRequest, db DB) error {
	_, err := SignNodeSig(c, db)
	if err != nil {
		return err
	}
	err = db.StoreContextRequest(c)
	if err != nil {
		return err
	}
	return nil
}

func validateAndSaveContextRequest(c *gripdata.ContextRequest, db DB) (*gripdata.Context, error) {
	ctx, err := validateNewContextRequest(c, db)
	if err != nil {
		return nil, err
	}
	err = signAndStoreContextRequest(c, db)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func sendToNewContextTarget(c *gripdata.ContextRequest, ctx *gripdata.Context, db DB) error {
	err := CreateNewSend(ctx, c.TargetNodeID, db)
	if err != nil {
		return errors.New("Failed to create send request")
	}
	err = CreateNewSend(c, c.TargetNodeID, db)
	if err != nil {
		return errors.New("Failed to create send request")
	}
	return err
}

func sendRequestAndResponse(ct *gripdata.ContextRequest, target []byte, db DB) error {
	err := CreateNewSend(ct, target, db)
	if err != nil {
		return err
	}
	rsp := db.GetContextResponse(ct.ContextDig, ct.TargetNodeID)
	if rsp != nil {
		//Send request and response to new node
		err = CreateNewSend(rsp, target, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func reciprocateNewContextData(c *gripdata.ContextRequest, ct *gripdata.ContextRequest, db DB) error {
	//Send the request and response to the new target node
	err := sendRequestAndResponse(ct, c.TargetNodeID, db)
	if err != nil {
		return err
	}
	//Send new request to other node
	err = CreateNewSend(c, ct.TargetNodeID, db)
	if err != nil {
		return err
	}
	return nil
}

func sendNewContextRequestToOthers(c *gripdata.ContextRequest, ctx *gripdata.Context, db DB) error {
	//Send to the creator
	err := CreateNewSend(c, ctx.NodeID, db)
	if err != nil {
		return fmt.Errorf("Failed to create send request: %s", err)
	}
	//Reciprocate data with other nodes participating in the context
	clr := db.GetContextRequests(c.ContextDig)
	for _, ct := range clr {
		err = reciprocateNewContextData(c, ct, db)
		if err != nil {
			return fmt.Errorf("Failed to reciprocate context data: %s", err)
		}
	}
	return nil
}

//NewContextRequest request a new node participate in
//a context
func NewContextRequest(c *gripdata.ContextRequest, db DB) error {
	ctx, err := validateAndSaveContextRequest(c, db)
	if err != nil {
		return err
	}
	err = sendToNewContextTarget(c, ctx, db)
	if err != nil {
		return err
	}
	err = sendNewContextRequestToOthers(c, ctx, db)
	if err != nil {
		return err
	}
	return nil
}

//SendToAllContextRequests sends data to all nodes that have a valid request for
//a context
func SendToAllContextRequests(c gripcrypto.SignInf, cid []byte, db DB) error {
	//Send to all with requests
	clr := db.GetContextRequests(cid)
	for _, ct := range clr {
		err := CreateNewSend(c, ct.TargetNodeID, db)
		if err != nil {
			return errors.New("Failed to create send request")
		}
	}
	return nil
}
