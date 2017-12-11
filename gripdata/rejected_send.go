package gripdata

//RejectedSendData Data that was not accepted by the
//target node
type RejectedSendData struct {
	TargetID  []byte //The node to send the data to
	Dig       []byte //The digest of the data to send
	Timestamp uint64 //The time this was created
	Message   string
}
