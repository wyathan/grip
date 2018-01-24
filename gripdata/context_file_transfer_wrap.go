package gripdata

//ContextFileTransferWrap wrapper for ContextFileTransfer to
//keep local information needed
type ContextFileTransferWrap struct {
	ContextFileTransfer *ContextFileTransfer
	ContextFile         *ContextFile
	Size                int64
	RetreivedOn         int64
}
