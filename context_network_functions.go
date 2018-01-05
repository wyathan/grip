package grip

import (
	"bytes"
	"errors"
	"log"
	"os"

	"github.com/wyathan/grip/gripdata"
)

//IncomingContext process a new incoming network
func IncomingContext(c *gripdata.Context, db NodeNetAccountContextdb) error {
	_, pr := db.GetPrivateNodeData()
	if bytes.Equal(pr.ID, c.NodeID) {
		//I created this.  Ignore it
		return nil
	}
	err := VerifyNodeSig(c, db)
	if err != nil {
		return err
	}
	err = db.StoreContext(c)
	if err != nil {
		return err
	}
	return nil
}

//IncomingContextRequest process new incoming ContextRequest
func IncomingContextRequest(c *gripdata.ContextRequest, db NodeNetAccountContextdb) error {
	//Check account is valide
	_, pr := db.GetPrivateNodeData()
	if bytes.Equal(pr.ID, c.NodeID) {
		//I created this.  Ignore it
		return nil
	}
	err := VerifyNodeSig(c, db)
	if err != nil {
		return err
	}
	//See if there's a duplicate request already from another node
	rqst := db.GetContextRequest(c.ContextDig, c.TargetNodeID)
	if rqst != nil {
		return errors.New("There is already a ContextRequest for this node")
	}
	//Check if requester can request!
	ctx := db.GetContext(c.ContextDig)
	if ctx == nil {
		return errors.New("Context was not found")
	}
	if !bytes.Equal(ctx.NodeID, c.NodeID) {
		//Check if the node was granted permission
		req := db.GetContextRequest(c.ContextDig, c.NodeID)
		rsp := db.GetContextResponse(c.ContextDig, c.NodeID)
		if req == nil || rsp == nil {
			return errors.New("Missing request/response from source node")
		}
		if !(req.AllowContextNode && rsp.ContextNode) {
			return errors.New("Source node does not have permission add nodes")
		}
		//Make sure that request creator is not escilating permissions for new node
		if !CheckNonEscalating(req, c) {
			return errors.New("Cannot escalate permissions")
		}
	}
	err = db.StoreContextRequest(c)
	if err != nil {
		return err
	}
	if bytes.Equal(pr.ID, c.TargetNodeID) {
		//This is for me
		if !IsAccountEnabled(c, db) {
			return errors.New("Sorry, this account is not enabled")
		}
		if pr.AutoContextResponse {
			a := GetNodeAccount(c.NodeID, db)
			if a.NumberContexts >= a.MaxContexts {
				return errors.New("Too many contexts for account")
			}

			//Create context response
			var r gripdata.ContextResponse
			r.CacheMode = a.AllowCacheMode & c.CacheMode
			r.ContextDig = c.ContextDig
			r.ContextSource = a.AllowContextSource
			r.ContextNode = a.AllowContextNode
			r.FullRepo = a.AllowFullRepo
			r.NewLogin = a.AllowNewLogin
			err = NewContextResponse(&r, db)
			if err != nil {
				return err
			}
		}
	}
	//Forward along to new node with the context
	//If me it won't send it
	err = CreateNewSend(ctx, c.TargetNodeID, db)
	if err != nil {
		return err
	}
	err = CreateNewSend(c, c.TargetNodeID, db)
	if err != nil {
		return err
	}
	//Forward to nodes participating in the context
	clr := db.GetContextRequests(c.ContextDig)
	for _, ct := range clr {
		//Also forward other context data to the target node, so if they
		//accept they can forward the acceptance along to other nodes
		err = CreateNewSend(ct, c.TargetNodeID, db)
		if err != nil {
			return err
		}
		rsp := db.GetContextResponse(ct.ContextDig, ct.TargetNodeID)
		if rsp != nil {
			//Send request and response to new node
			err = CreateNewSend(rsp, c.TargetNodeID, db)
			if err != nil {
				return err
			}
		}
		//Send new request to other nodes
		err = CreateNewSend(c, ct.TargetNodeID, db)
		if err != nil {
			return err
		}
	}
	return nil
}

//IncomingContextResponse processes incoming context responses.  Permission is checked
//on requests, so we can just accept all responses.  Note we don't check for the request
//first just in case we get them out of order.  However, to participate in the context
//a valid request must match the response.
func IncomingContextResponse(c *gripdata.ContextResponse, db NodeNetAccountContextdb) error {
	_, pr := db.GetPrivateNodeData()
	if bytes.Equal(pr.ID, c.TargetNodeID) {
		//I created this.  Ignore it
		return nil
	}
	err := VerifyNodeSig(c, db)
	if err != nil {
		return err
	}
	//Save to the database
	err = db.StoreContextResponse(c)
	if err != nil {
		return err
	}
	//Forward to context creator
	ctx := db.GetContext(c.ContextDig)
	if ctx == nil {
		return errors.New("Context not found")
	}
	err = CreateNewSend(c, ctx.NodeID, db)
	if err != nil {
		return err
	}
	//Forward to nodes participating in the context
	err = SendToAllContextRequests(c, c.ContextDig, db)
	if err != nil {
		return err
	}
	return nil
}

//IncomingContextFile process an incoming context file.  Check permission and forward
//to participating nodes
func IncomingContextFile(c *gripdata.ContextFile, db NodeNetAccountContextdb) error {
	_, pr := db.GetPrivateNodeData()
	if bytes.Equal(pr.ID, c.NodeID) {
		//I created this.  Ignore it
		return nil
	}
	//Check the signature of the context file
	err := VerifyNodeSig(c, db)
	if err != nil {
		return err
	}
	ctx := db.GetContext(c.Context)
	if !IsIfValidContextSource(c.NodeID, ctx, db) {
		db.StoreVeryBadContextFile(c)
		log.Printf("Incoming ContextFile without permission: %s", c.Dig)
		return errors.New("Not valid source node ")
	}
	if !IsContextFileDepsOk(c, db) {
		return errors.New("Dependency problems were found")
	}
	//Get the account for the creating node.
	a := GetNodeAccount(c.NodeID, db)
	if !((bytes.Equal(ctx.NodeID, c.NodeID) || a.AllowContextSource) && a.Enabled) {
		return errors.New("Account does not allow context source")
	}
	//You are here
	//Check file size
	st, err2 := os.Stat(c.Path)
	if err2 != nil {
		return err2
	}
	//Check max file storage
	err = db.CheckUpdateStorageUsed(a, uint64(st.Size()))
	if err != nil {
		return err
	}
	_, err = db.StoreContextFile(c)
	if err != nil {
		return err
	}
	//Forward to other participants
	err = SendToAllContextParticipants(c, c.Context, db)
	if err != nil {
		return err
	}
	return nil
}
