package grip

import (
	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
	"github.com/wyathan/grip/griperrors"
)

type FTProc func(c *gripdata.ContextFileTransfer, db DB) (bool, error)

func FileTransferProc(f FTProc) SProc {
	return func(c gripcrypto.SignInf, db DB) (bool, error) {
		switch v := c.(type) {
		case *gripdata.ContextFileTransfer:
			return f(v, db)
		default:
			return false, griperrors.WrongType
		}
	}
}

func StoreContextFileTransfer(c *gripdata.ContextFileTransfer, db DB) (bool, error) {
	_, err := db.StoreContextFileTransfer(c)
	return true, err
}

func IncomingFileTransfer(c *gripdata.ContextFileTransfer, db DB) error {
	var s SProcChain
	s.Push(LocallyCreated)
	s.Push(VerifyNodeSig)
	s.Push(FileTransferProc(StoreContextFileTransfer))
	s.Push(PrepSendToAllContextParticipants(c.Context))
	_, err := s.F(c, db)
	return err
}
