package gripdata

//SendData Indicates this data shohuld be sent to this node when
//possible
type SendData struct {
	TargetID  []byte //The node to send the data to
	Dig       []byte //The digest of the data to send
	Timestamp uint64 //The time this was created
	TypeName  string //The struct type
}
