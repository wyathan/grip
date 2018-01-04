package grip

import "github.com/wyathan/grip/gripdata"

//IsCoveredBySnapshots determines if the latest version of
//the repository would still be buidable from later snapshots
//if we removed this file.
func IsCoveredBySnapshots(f *gripdata.ContextFile, db Contextdb) bool {
	lst := db.GetAllThatDependOn(f.Context, f.DataDepDig)
	if 0 == len(lst) {
		return false
	}
	ok := true
	for _, v := range lst {
		if !v.Snapshot {
			if !IsCoveredBySnapshots(&v, db) {
				return false
			}
		}
	}
	return ok
}

//FreeSpace frees needed space for the context listed
func FreeSpace(context []byte, needed uint64, db Contextdb) error {
	if needed > 0 {
		//Remove from non-index files first

	}
	return nil
}
