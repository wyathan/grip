package griptests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/wyathan/grip/gripdata"
)

func showType(v interface{}) {
	to := reflect.TypeOf(v)
	fmt.Println(to.String())
}

func TestReflection(t *testing.T) {
	var n gripdata.Node
	var a gripdata.Account
	showType(n)
	showType(a)
}
