package grip

import (
	"errors"

	"github.com/wyathan/grip/gripdata"
)

func signAndStoreContextResponse(c *gripdata.ContextResponse, db NodeNetContextdb) error {
	err := SignNodeSig(c, db)
	if err != nil {
		return err
	}
	err = db.StoreContextResponse(c)
	if err != nil {
		return err
	}
	return nil
}

func validateAndSaveContextResponse(c *gripdata.ContextResponse, db NodeNetContextdb) (*gripdata.Context, error) {
	//Make sure there was a request
	myn, _ := db.GetPrivateNodeData()
	req := db.GetContextRequest(c.ContextDig, myn.ID)
	if req == nil {
		return nil, errors.New("You have to have a request first")
	}
	ctx := db.GetContext(c.ContextDig)
	if ctx == nil {
		return nil, errors.New("Context is not found")
	}
	err := signAndStoreContextResponse(c, db)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

//NewContextResponse we have decided to participate in a context
func NewContextResponse(c *gripdata.ContextResponse, db NodeNetContextdb) error {
	ctx, err := validateAndSaveContextResponse(c, db)
	if err != nil {
		return err
	}
	err = CreateNewSend(c, ctx.NodeID, db)
	if err != nil {
		return errors.New("Failed to create send request")
	}
	//Send to all with requests
	err = SendToAllContextRequests(c, c.ContextDig, db)
	if err != nil {
		return err
	}
	return nil
}
