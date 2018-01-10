package grip

import (
	"errors"
	"log"

	"github.com/wyathan/grip/gripdata"
)

//FreeSpace frees needed space for the context listed
func FreeSpace(c *gripdata.Context, fsize uint64, db DB) error {
	if fsize > 0 {
		context := c.Dig
		//Remove snapshots covered by other snapshots first.
		//Snapshots contain no new changes themselves.  They just contain
		//complete files that are the result of all previous changes, so
		//they can bee removed so long as they are covered by a later
		//snapshot without loosing information.  They probably take
		//the most space as well
		a := GetNodeAccount(c.NodeID, db)
		if a == nil {
			return errors.New("No Account found for context creator")
		}
		cs := db.GetCoveredSnapshots(context)
		if len(cs) > 0 {
			//The last elements will be the deepest of the largest size, get rid of them first
			for c := 0; c < len(cs) && fsize > a.Free(); c++ {
				if cs[c].Size > 0 {
					a, _ = DeleteContextFileWrap(cs[c], a, db)
				}
			}
		}
		loopmore := true
		for fsize > a.Free() && loopmore {
			loopmore = false
			//Booo.. we need to loose change history to save space.
			//Try removing from non-index history first
			cs := db.GetContextLeaves(context, true, false)
			for c := 0; c < len(cs) && fsize > a.Free(); c++ {
				if cs[c].Size > 0 {
					a, _ = DeleteContextFileWrap(cs[c], a, db)
					loopmore = true
				}
			}
			if fsize > a.Free() {
				//Super boo!!  We have to loose history from the index!
				cs := db.GetContextLeaves(context, true, true)
				for c := 0; c < len(cs) && fsize > a.Free(); c++ {
					if cs[c].Size > 0 {
						a, _ = DeleteContextFileWrap(cs[c], a, db)
						loopmore = true
					}
				}
			}
		}
		loopmore = true
		for fsize > a.Free() && loopmore {
			loopmore = false
			//BOO!!  sniff sniff.  We have to remove changes needed to build
			//the current versions.
			cs := db.GetContextLeaves(context, false, false)
			for c := 0; c < len(cs) && fsize > a.Free(); c++ {
				if cs[c].Size > 0 {
					a, _ = DeleteContextFileWrap(cs[c], a, db)
					loopmore = true
				}
			}
		}
		if fsize > a.Free() {
			//This is catastrophic, we'd have to remove changes needed to build
			//the current version of the index.  If we do this we're useless to
			//new nodes.
			db.SetAccountMessage(a, "Out of space, index too large")
			return errors.New("Out of space, index too large")
		}
	}
	return nil
}

func DeleteContextFileWrap(c *gripdata.ContextFileWrap, a *gripdata.Account, db DB) (*gripdata.Account, error) {
	a, err := db.FreeStorageUsed(a, uint64(c.Size))
	if err != nil {
		log.Printf("Failed to free disk space for account, %s", err)
		return a, err
	}
	//TODO: Acctually delete the file
	return a, nil
}
