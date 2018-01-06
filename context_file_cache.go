package grip

import (
	"github.com/wyathan/grip/gripdata"
)

//FreeSpace frees needed space for the context listed
func FreeSpace(context []byte, needed uint64, db Contextdb) error {
	if needed > 0 {
		//Remove snapshots covered by other snapshots first.
		//Snapshots contain no new changes themselves.  They just contain
		//complete files that are the result of all previous changes, so
		//they can bee removed so long as they are covered by a later
		//snapshot without loosing information.  They probably take
		//the most space as well
		cs := db.GetCoveredSnapshots(context)
		if len(cs) > 0 {
			//The last elements will be the deepest of the largest size, get rid of them first
			for c := 0; c < len(cs) && needed > 0; c++ {
				if cs[c].Size > 0 {
					err := DeleteContextFileWrap(cs[c], db)
					if err == nil {
						needed -= uint64(cs[c].Size)
					}
				}
			}
		}
		if needed > 0 {
			//Booo.. we need to loose change history to save space.
			//Try removing from non-index history first
			cs := db.GetContextLeaves(context, true, false)
			for c := 0; c < len(cs) && needed > 0; c++ {
				if cs[c].Size > 0 {
					err := DeleteContextFileWrap(cs[c], db)
					if err == nil {
						needed -= uint64(cs[c].Size)
					}
				}
			}
		}
		if needed > 0 {
			//Super boo!!  We have to loose history from the index!
			cs := db.GetContextLeaves(context, true, true)
			for c := 0; c < len(cs) && needed > 0; c++ {
				if cs[c].Size > 0 {
					err := DeleteContextFileWrap(cs[c], db)
					if err == nil {
						needed -= uint64(cs[c].Size)
					}
				}
			}
		}
		if needed > 0 {
			//BOO!!  sniff sniff.  We have to remove changes needed to build
			//the current versions.
			cs := db.GetContextLeaves(context, false, false)
			for c := 0; c < len(cs) && needed > 0; c++ {
				if cs[c].Size > 0 {
					err := DeleteContextFileWrap(cs[c], db)
					if err == nil {
						needed -= uint64(cs[c].Size)
					}
				}
			}
		}
		if needed > 0 {
			//This is catastrophic, we'd have to remove changes needed to build
			//the current version of the index.  If we do this we're useless to
			//new nodes.
			//TODO: here
		}
	}
	return nil
}

func DeleteContextFileWrap(c *gripdata.ContextFileWrap, db Contextdb) error {
	//TODO: do this
	return nil
}
