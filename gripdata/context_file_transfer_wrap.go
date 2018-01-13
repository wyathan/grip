package gripdata

//ContextFileTransferWrap wrapper for ContextFileTransfer to
//keep local information needed
type ContextFileTransferWrap struct {
	ContextFileTransfer *ContextFileTransfer
	Size                int64
	Path                string
	RetreivedOn         int64
}
