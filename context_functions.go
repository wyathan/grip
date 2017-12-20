package grip

import (
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
