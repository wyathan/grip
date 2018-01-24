package gripdata

//DeletedContextFile a record of deleted ContextFile's should be
//create atomically on delete of ContextFile by database
type DeletedContextFile struct {
	Context    []byte
	Dig        []byte
	DataDepDig []byte
}
