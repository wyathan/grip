package grip

import (
	"bytes"
	"errors"
	"log"
	"os"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

func doesContextFielHaveLoop(head []byte, dep *gripdata.ContextFile, db NodeNetContextdb) bool {
	for _, dr := range dep.DependsOn {
		if isThereDepLoop(head, dr, db) {
			return true
		}
	}
	return false
}

//isThereDepLoop returns true if a dependency loop is found.
//It would be very impressive, but equally nasty if it occurs
func isThereDepLoop(head []byte, check []byte, db NodeNetContextdb) bool {
	if bytes.Equal(head, check) {
		return true
	}
	dep := db.GetContextFileByDepDataDig(check)
	if dep != nil {
		if doesContextFielHaveLoop(head, dep, db) {
			return true
		}
	}
	return false
}

func isDupInDeps(depdigs [][]byte, idx int) bool {
	//Check that no dependencies are listed twice
	for c1 := idx + 1; c1 < len(depdigs); c1++ {
		if bytes.Equal(depdigs[idx], depdigs[c1]) {
			log.Println("Someone was just dumb and listed duplicate dependencies.")
			return true
		}
	}
	return false
}

func areDepsSane(depdig []byte, chkdig []byte, db NodeNetContextdb) bool {
	//Check that the digest of dependencies and the file don't produce the
	//same value was one of the dependencies.  That would be very impressive!
	if bytes.Equal(depdig, chkdig) {
		log.Println("Something probably isn't working how you think it does.")
		return false
	}
	//Check that there are no loops in the dependencies
	if isThereDepLoop(depdig, chkdig, db) {
		log.Println("Something very bad happened.  A dependency loop was found.")
		return false
	}
	return true
}

//IsContextFileDepsOk returns true if the depenencies
//are ok and won't cause problems
func IsContextFileDepsOk(c *gripdata.ContextFile, db NodeNetContextdb) (r bool) {
	defer func() {
		if !r && c != nil {
			db.StoreVeryBadContextFile(c)
		}
	}()
	if c.DependsOn != nil {
		for c0 := 0; c0 < len(c.DependsOn); c0++ {
			if !areDepsSane(c.DataDepDig, c.DependsOn[c0], db) {
				return false
			}
			if isDupInDeps(c.DependsOn, c0) {
				return false
			}
		}
	}
	return true
}

func filterContextRequest(c *gripdata.ContextRequest, db NodeNetContextdb) *gripdata.ContextRequest {
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
func IsIfValidContextSource(nid []byte, ctx *gripdata.Context, db NodeNetContextdb) bool {
	if ctx == nil {
		return false
	}
	if bytes.Equal(ctx.NodeID, nid) {
		return true
	}
	req := db.GetContextRequest(ctx.Dig, nid)
	if req == nil {
		return false
	}
	rq := filterContextRequest(req, db)
	return rq != nil && rq.AllowContextSource
}

func sendToContextParticipant(c gripcrypto.SignInf, ct *gripdata.ContextRequest, db NodeNetContextdb) error {
	fr := filterContextRequest(ct, db)
	if fr != nil {
		err := CreateNewSend(c, ct.TargetNodeID, db)
		if err != nil {
			return errors.New("Failed to create send request")
		}
	}
	return nil
}

func findAndSendToContextParticipants(c gripcrypto.SignInf, ctxid []byte, db NodeNetContextdb) error {
	clr := db.GetContextRequests(ctxid)
	for _, ct := range clr {
		err := sendToContextParticipant(c, &ct, db)
		if err != nil {
			return err
		}
	}
	return nil
}

//SendToAllContextParticipants Send context data to all participants
func SendToAllContextParticipants(c gripcrypto.SignInf, ctxid []byte, db NodeNetContextdb) error {
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
	err = findAndSendToContextParticipants(c, ctxid, db)
	if err != nil {
		return err
	}
	return nil
}

func validateFile(c *gripdata.ContextFile) error {
	st, err := os.Stat(c.Path)
	if os.IsNotExist(err) {
		return errors.New("Path does not exist")
	}
	if (st.Mode() & os.ModeType) != 0 {
		return errors.New("Path is not a regular file")
	}
	return nil
}

func isAllowedToCreateFile(c *gripdata.ContextFile, db NodeNetContextdb) error {
	myn, _ := db.GetPrivateNodeData()
	ctx := db.GetContext(c.Context)
	if ctx == nil {
		return errors.New("Context is unknown")
	}
	if !IsIfValidContextSource(myn.ID, ctx, db) {
		return errors.New("You do not have context source permission")
	}
	return nil
}

func validateContextFileDepsAndSign(c *gripdata.ContextFile, db NodeNetContextdb) error {
	//Go ahead and sign so that we have valid digests
	err := SignNodeSig(c, db)
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
	return nil
}

func validateAndSignContextFile(c *gripdata.ContextFile, db NodeNetContextdb) error {
	err := validateFile(c)
	if err != nil {
		return err
	}
	err = isAllowedToCreateFile(c, db)
	if err != nil {
		return err
	}
	return validateContextFileDepsAndSign(c, db)
}

//NewContextFile add a new file for a context
func NewContextFile(c *gripdata.ContextFile, db NodeNetContextdb) error {
	err := validateAndSignContextFile(c, db)
	if err != nil {
		return err
	}
	err = db.StoreContextFile(c)
	if err != nil {
		return err
	}
	return SendToAllContextParticipants(c, c.Context, db)
}
