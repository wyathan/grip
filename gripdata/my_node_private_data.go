package gripdata

//MyNodePrivateData is private information about this node
//never shared with another.
type MyNodePrivateData struct {
	ID                []byte //Unique id for the node (digest of public key)
	BindAddress       string //where to bind listen port
	BindPort          uint32 //listen port to use
	PrivateKey        []byte
	PrivateMetaData   string //Extra private generic metadata
	AutoShareNodeInfo bool
	AutoShareMetaData string
}
