package grip

import (
	"github.com/wyathan/grip/gripcrypto"
)

//SProc a function that operates on a SignInf struct
type SProc func(s gripcrypto.SignInf, db DB) (bool, error)

//SProcChain a list of ordered functions to execute on SignInf
type SProcChain struct {
	funcs []sProcElm
}

type sProcElm struct {
	f SProc
	v gripcrypto.SignInf
}

//Push add another function to the list
func (p *SProcChain) Push(f SProc) {
	p.funcs = append(p.funcs, sProcElm{f, nil})
}

//PushV add another function to the list with a different value
func (p *SProcChain) PushV(f SProc, v gripcrypto.SignInf) {
	p.funcs = append(p.funcs, sProcElm{f, v})
}

//F runs a series of SProc functions
func (p *SProcChain) F(s gripcrypto.SignInf, db DB) (bool, error) {
	for _, f := range p.funcs {
		var err error
		var cont bool
		if f.v != nil {
			cont, err = f.f(f.v, db)
		} else {
			cont, err = f.f(s, db)
		}
		if !cont || err != nil {
			return cont, err
		}
	}
	return true, nil
}
