package grip

import (
	"bytes"
	"errors"
	"log"
	"os"
	"time"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

//NewContext signs/stores a new context
func NewContext(c *gripdata.Context, db NodeContextdb) error {
	//Get my node id data
	n, pr := db.GetPrivateNodeData()
	c.SetNodeID(n.ID)
	c.CreatedOn = uint64(time.Now().UnixNano())
	err := gripcrypto.Sign(c, pr.PrivateKey)
	if err != nil {
		return err
	}
	err = db.StoreContext(c)
	if err != nil {
		return err
	}
	return nil
}

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

//NewContextRequest request a new node participate in
//a context
func NewContextRequest(c *gripdata.ContextRequest, db NodeNetConextdb) error {
	tid := db.GetNode(c.TargetNodeID)
	if tid == nil {
		return errors.New("Unknown target node")
	}
	ctx := db.GetContext(c.ContextDig)
	if ctx == nil {
		return errors.New("Context is not found")
	}
	myn, _ := db.GetPrivateNodeData()
	if !bytes.Equal(ctx.NodeID, myn.ID) {
		//See if we have permission to add others
		myreq := db.GetContextRequest(ctx.Dig, myn.ID)
		myrsp := db.GetContextResponse(ctx.Dig, myn.ID)
		if !(myreq.AllowContextNode && myrsp.ContextNode) {
			return errors.New("Not allowed to request new nodes for context")
		}
		if !CheckNonEscalating(myreq, c) {
			return errors.New("You cannot escilate permissions for a new node")
		}
	}

	rqst := db.GetContextRequest(c.ContextDig, c.TargetNodeID)
	if rqst != nil {
		return errors.New("There is already a ContextRequest for this node")
	}
	err := SignNodeSig(c, db)
	if err != nil {
		return err
	}
	err = db.StoreContextRequest(c)
	if err != nil {
		return err
	}
	//Send to the target node
	err = CreateNewSend(ctx, c.TargetNodeID, db)
	if err != nil {
		return errors.New("Failed to create send request")
	}
	err = CreateNewSend(c, c.TargetNodeID, db)
	if err != nil {
		return errors.New("Failed to create send request")
	}
	//Send to context creator
	err = CreateNewSend(c, ctx.NodeID, db)
	if err != nil {
		return errors.New("Failed to create send request")
	}
	//Forward to nodes participating in the context
	clr := db.GetContextRequests(c.ContextDig)
	for _, ct := range clr {
		//Also forward other context data to the target node, so if they
		//accept they can forward the acceptance along to other nodes
		err = CreateNewSend(&ct, c.TargetNodeID, db)
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

//NewContextResponse we have decided to participate in a context
func NewContextResponse(c *gripdata.ContextResponse, db NodeNetConextdb) error {
	//Make sure there was a request
	myn, _ := db.GetPrivateNodeData()
	req := db.GetContextRequest(c.ContextDig, myn.ID)
	if req == nil {
		return errors.New("You have to have a request first")
	}
	err := SignNodeSig(c, db)
	if err != nil {
		return err
	}
	err = db.StoreContextResponse(c)
	if err != nil {
		return err
	}
	//NOTE: Do NOT use SendToAllContextParticipants because we
	//want to send to all nodes with requests, not just the ones
	//that also responded
	//Send to the context owner
	ctx := db.GetContext(c.ContextDig)
	if ctx == nil {
		return errors.New("Context is not found")
	}
	err = CreateNewSend(c, ctx.NodeID, db)
	if err != nil {
		return errors.New("Failed to create send request")
	}
	//Send to all with requests
	clr := db.GetContextRequests(c.ContextDig)
	for _, ct := range clr {
		err = CreateNewSend(c, ct.TargetNodeID, db)
		if err != nil {
			return errors.New("Failed to create send request")
		}
	}
	return nil
}

//isThereDepLoop returns true if a dependency loop is found.
//It would be very impressive, but equally nasty if it occurs
func isThereDepLoop(head []byte, check []byte, db NodeNetConextdb) bool {
	if bytes.Equal(head, check) {
		return true
	}
	dep := db.GetContextFileByDepDataDig(check)
	if dep != nil {
		if dep.DependsOn != nil {
			for _, dr := range dep.DependsOn {
				if isThereDepLoop(head, dr, db) {
					return true
				}
			}
		}
	}
	return false
}

//IsContextFileDepsOk returns true if the depenencies
//are ok and won't cause problems
func IsContextFileDepsOk(c *gripdata.ContextFile, db NodeNetConextdb) (r bool) {
	defer func() {
		if !r && c != nil {
			db.StoreVeryBadContextFile(c)
		}
	}()
	if c.DependsOn != nil {
		for c0 := 0; c0 < len(c.DependsOn); c0++ {
			//Check that the digest of dependencies and the file don't produce the
			//same value was one of the dependencies.  That would be very impressive!
			if bytes.Equal(c.DataDepDig, c.DependsOn[c0]) {
				log.Println("Something probably isn't working how you think it does.")
				return false
			}
			//Check that no dependencies are listed twice
			for c1 := c0 + 1; c1 < len(c.DependsOn); c1++ {
				if bytes.Equal(c.DependsOn[c0], c.DependsOn[c1]) {
					log.Println("Someone was just dumb and listed duplicate dependencies.")
					return false
				}
			}
			//Check that there are no loops in the dependencies
			if isThereDepLoop(c.DataDepDig, c.DependsOn[c0], db) {
				log.Println("Something very bad happened.  A dependency loop was found.")
				return false
			}

		}
	}
	return true
}

func filterContextRequest(c *gripdata.ContextRequest, db NodeNetConextdb) *gripdata.ContextRequest {
	if c == nil {
		return nil
	}
	rsp := db.GetContextResponse(c.ContextDig, c.TargetNodeID)
	if rsp == nil {
		return nil
	}
	r := *c
	r.CacheMode &= rsp.CacheMode
	r.AllowContextNode = r.AllowContextNode && rsp.ContextNode
	r.AllowContextSource = r.AllowContextSource && rsp.ContextSource
	r.AllowNewLogin = r.AllowNewLogin && rsp.NewLogin
	return &r
}

//IsIfValidContextSource check if nid or the
func IsIfValidContextSource(nid []byte, ctx *gripdata.Context, req *gripdata.ContextRequest, db NodeNetConextdb) bool {
	if ctx == nil {
		return false
	}
	if req != nil {
		nid = req.TargetNodeID
	}
	if req == nil {
		req = db.GetContextRequest(ctx.Dig, nid)
	}
	if req == nil {
		return false
	}
	if bytes.Equal(ctx.NodeID, nid) {
		return true
	}
	rq := filterContextRequest(req, db)
	return rq != nil && rq.AllowContextSource
}

//SendToAllContextParticipants Send context data to all participants
func SendToAllContextParticipants(c gripcrypto.SignInf, ctxid []byte, db NodeNetConextdb) error {
	//Send to the context owner
	ctx := db.GetContext(ctxid)
	if ctx == nil {
		return errors.New("Context is not found")
	}
	err := CreateNewSend(c, ctx.NodeID, db)
	if err != nil {
		return errors.New("Failed to create send request")
	}
	//Send to all with responses
	clr := db.GetContextRequests(ctxid)
	for _, ct := range clr {
		fr := filterContextRequest(&ct, db)
		if fr != nil {
			err = CreateNewSend(c, ct.TargetNodeID, db)
			if err != nil {
				return errors.New("Failed to create send request")
			}
		}
	}
	return nil

}

//NewContextFile add a new file for a context
func NewContextFile(c *gripdata.ContextFile, db NodeNetConextdb) error {
	st, err := os.Stat(c.Path)
	if os.IsNotExist(err) {
		return errors.New("Path does not exist")
	}
	if (st.Mode() & os.ModeType) != 0 {
		return errors.New("Path is not a regular file")
	}
	myn, _ := db.GetPrivateNodeData()
	ctx := db.GetContext(c.Context)
	if ctx == nil {
		return errors.New("Context is unknown")
	}
	if !IsIfValidContextSource(myn.ID, ctx, nil, db) {
		return errors.New("You do not have context source permission")
	}
	//Go ahead and sign so that we have valid digests
	err = SignNodeSig(c, db)
	if err != nil {
		return err
	}
	dd := db.GetContextFileByDepDataDig(c.DataDepDig)
	if dd != nil {
		return errors.New("We already have this file")
	}
	//Check for dependency problems
	if !IsContextFileDepsOk(c, db) {
		return errors.New("There were dependency problems")
	}
	err = db.StoreContextFile(c)
	if err != nil {
		return err
	}
	err = SendToAllContextParticipants(c, c.Context, db)
	if err != nil {
		return err
	}
	return nil
}
