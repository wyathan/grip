package grip

import (
	"encoding/base64"
	"log"
	"math/rand"
	"time"

	"github.com/wyathan/grip/gripdata"
)

func (s *SocketController) connectToNodeEphemera(c *gripdata.NodeEphemera) {
	if s.DB.CanNodeEphemeraGoPending(c.ID) {
		n := s.DB.GetNode(c.ID)
		con, err := s.S.ConnectTo(n)
		if con != nil && err == nil {
			s.buildConnectionController(con, false)
		} else {
			if con != nil {
				con.Close()
			}
			log.Print("Connection error\n")
			nt := uint64(time.Now().UnixNano())
			s.DB.SetNodeEphemeraNextConnection(c.ID, nt, nt+uint64(TRYCONNECTAGAINAFTERFAIL.Nanoseconds()))
		}
	}
}

func (s *SocketController) canAttempt(attempted *map[string]uint64, id []byte, nowtime uint64) bool {
	sid := base64.StdEncoding.EncodeToString(id)
	oldenough := nowtime - uint64(WAITUNTILCONNECTAGAIN)
	if (*attempted)[sid] < oldenough {
		(*attempted)[sid] = nowtime
		return true
	}
	return false
}
func (s *SocketController) connectToNodesWithSendData(attempted *map[string]uint64) {
	nt := uint64(time.Now().UnixNano())
	nm := s.numberConnections()
	if nm <= MAXCONNECTIONS {
		cl := s.DB.GetConnectableNodesWithSendData(MAXCONNECTIONATTEMPTS, nt)
		for _, c := range cl {
			if s.canAttempt(attempted, c.ID, nt) {
				s.connectToNodeEphemera(&c)
			}
		}
	}
}
func (s *SocketController) connectToNodesWithShareNodeKey(attempted *map[string]uint64) {
	nt := uint64(time.Now().UnixNano())
	nm := s.numberConnections()
	if nm <= MAXCONNECTIONS {
		cl := s.DB.GetConnectableNodesWithShareNodeKey(MAXCONNECTIONATTEMPTS, nt)
		for _, c := range cl {
			if s.canAttempt(attempted, c.ID, nt) {
				s.connectToNodeEphemera(&c)
			}
		}
	}
}
func (s *SocketController) connectToNodesWithUseShareKey(attempted *map[string]uint64) {
	nt := uint64(time.Now().UnixNano())
	nm := s.numberConnections()
	if nm <= MAXCONNECTIONS {
		cl := s.DB.GetConnectableUseShareKeyNodes(MAXCONNECTIONATTEMPTS, nt)
		for _, c := range cl {
			if s.canAttempt(attempted, c.ID, nt) {
				s.connectToNodeEphemera(&c)
			}
		}
	}
}
func (s *SocketController) connectToAnyNodes(attempted *map[string]uint64) {
	nt := uint64(time.Now().UnixNano())
	nm := s.numberConnections()
	if nm <= MAXCONNECTIONS {
		cl := s.DB.GetConnectableAny(MAXCONNECTIONATTEMPTS, nt)
		//Select one at random
		ln := len(cl)
		if ln > 0 {
			idx := rand.Intn(len(cl))
			if s.canAttempt(attempted, cl[idx].ID, nt) {
				log.Printf("Attempt random connection")
				s.connectToNodeEphemera(&cl[idx])
			}
		}
	}
}
func (s *SocketController) connectRoutine() {
	attempted := make(map[string]uint64)
	for !s.Done {
		s.connectToNodesWithSendData(&attempted)
		s.connectToNodesWithShareNodeKey(&attempted)
		s.connectToNodesWithUseShareKey(&attempted)
		time.Sleep(CONNECTROUTINESLEEP)
		s.connectToAnyNodes(&attempted)
		s.checkConnections()
	}
}
