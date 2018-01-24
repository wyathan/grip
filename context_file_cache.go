package grip

import (
	"errors"
	"log"
	"os"

	"github.com/wyathan/grip/gripdata"
	"github.com/wyathan/grip/griperrors"
)

func deleteCoveredSnapshots(a *gripdata.Account, c *gripdata.Context, fsize uint64, db DB) error {
	cs := db.GetCoveredSnapshots(c.Dig)
	var err error
	for c := 0; c < len(cs) && fsize > a.Free(); c++ {
		a, err = deleteContextFileWrap(cs[c], a, db)
		if err != nil {
			log.Printf("Could not delete covered snapshots: %s", err)
			return err
		}
	}
	return nil
}

func deleteHistory(a *gripdata.Account, c *gripdata.Context, fsize uint64, cvr bool, idx bool, db DB) (bool, error) {
	loopmore := false
	var err error
	cs := db.GetContextLeaves(c.Dig, cvr, idx)
	for c := 0; c < len(cs) && fsize > a.Free(); c++ {
		a, err = deleteContextFileWrap(cs[c], a, db)
		if err != nil {
			return false, err
		}
		loopmore = true
	}
	return loopmore, nil
}

func deleteCoveredHistory(a *gripdata.Account, c *gripdata.Context, fsize uint64, db DB) error {
	loopmore := true
	var err error
	for fsize > a.Free() && loopmore {
		//Try to delete covered, non-index history first
		loopmore, err = deleteHistory(a, c, fsize, true, false, db)
		if err != nil {
			return err
		}
		if fsize > a.Free() {
			//Super boo!!  We have to loose history from the index!
			loopmore, err = deleteHistory(a, c, fsize, true, true, db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteNonCoveredHistory(a *gripdata.Account, c *gripdata.Context, fsize uint64, db DB) error {
	loopmore := true
	var err error
	for fsize > a.Free() && loopmore {
		loopmore = false
		cs := db.GetContextLeaves(c.Dig, false, false)
		for c := 0; c < len(cs) && fsize > a.Free(); c++ {
			a, err = deleteContextFileWrap(cs[c], a, db)
			loopmore = true
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//FreeSpace frees needed space for the context listed
func FreeSpace(c *gripdata.Context, fsize uint64, db DB) error {
	if fsize > 0 {
		a := GetNodeAccount(c.NodeID, db)
		if a == nil {
			return griperrors.AccountNotFound
		}
		err := deleteCoveredSnapshots(a, c, fsize, db)
		if err != nil {
			return err
		}
		err = deleteCoveredHistory(a, c, fsize, db)
		if err != nil {
			return err
		}
		err = deleteNonCoveredHistory(a, c, fsize, db)
		if err != nil {
			return err
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

func deleteContextFileWrap(c *gripdata.ContextFileWrap, a *gripdata.Account, db DB) (*gripdata.Account, error) {
	err := db.DeleteContextFile(c)
	if err != nil {
		return a, err
	}
	err = os.Remove(c.ContextFile.GetPath())
	if err != nil {
		return a, err
	}
	a, err = db.FreeStorageUsed(a, c.ContextFile.Size)
	if err != nil {
		log.Printf("Failed to free disk space for account, %s", err)
		return a, err
	}
	return a, nil
}
