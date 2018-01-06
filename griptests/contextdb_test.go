package griptests

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/wyathan/grip/gripdata"
)

func CreateTestContextFile(cid string, dig string, dd string) *gripdata.ContextFile {
	var c gripdata.ContextFile
	c.Context = []byte(cid)
	c.Dig = []byte(dig)
	c.DataDepDig = []byte(dd)
	return &c
}

func TestContextFileWrap(t *testing.T) {
	go func() {
		// http://localhost:6060/debug/pprof/goroutine?debug=2
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	//                 15
	//                  |
	//             10   7    12     11
	//             |      \   /      |
	//             9        6        8
	//              \      / \      /
	//                 4         5         13      14
	//                 |          \       /
	//                 3              2
	//                                |
	//                                1
	db := NewTestDB()

	n1 := CreateTestContextFile("context", "1", "dd1")
	n1.Snapshot = true
	w1, err := db.StoreContextFile(n1)
	if err != nil || !(w1.Depth == 0 && w1.Head && w1.Leaf) {
		t.Error()
	}

	n2 := CreateTestContextFile("context", "2", "dd2").PushDep(n1.DataDepDig)
	w2, err2 := db.StoreContextFile(n2)
	if err2 != nil || !(w2.Depth == 0 && w2.Head && !w2.Leaf) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 1 && !w1.Head && w1.Leaf) {
		t.Error()
	}

	n3 := CreateTestContextFile("context", "3", "dd3")
	n3.Snapshot = true
	w3, err3 := db.StoreContextFile(n3)
	if err3 != nil || !(w3.Depth == 0 && w3.Head && w3.Leaf) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 0 && w2.Head && !w2.Leaf) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 1 && !w1.Head && w1.Leaf) {
		t.Error()
	}

	n4 := CreateTestContextFile("context", "4", "dd4").PushDep(n3.DataDepDig)
	w4, err4 := db.StoreContextFile(n4)
	if err4 != nil || !(w4.Depth == 0 && w4.Head && !w4.Leaf) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 1 && !w3.Head && w3.Leaf) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 0 && w2.Head && !w2.Leaf) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 1 && !w1.Head && w1.Leaf) {
		t.Error()
	}

	n5 := CreateTestContextFile("context", "5", "dd5").PushDep(n2.DataDepDig)
	w5, err5 := db.StoreContextFile(n5)
	if err5 != nil || !(w5.Depth == 0 && w5.Head && !w5.Leaf) {
		if err5 != nil {
			fmt.Printf("Error is not nil %s", err5)
		}
		if w5.Depth != 0 {
			fmt.Print("Depth is not 0")
		}
		if !w5.Head {
			fmt.Print("Not head")
		}
		if w5.Leaf {
			fmt.Print("Is leaf")
		}
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 0 && w4.Head && !w4.Leaf) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 1 && !w3.Head && w3.Leaf) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 1 && !w2.Head && !w2.Leaf) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 2 && !w1.Head && w1.Leaf) {
		log.Printf("depth %d", w1.Depth)
		log.Printf("head %t", w1.Head)
		log.Printf("leaf %t", w1.Leaf)
		t.Error()
	}

	// 6: 4, 5, 3, 2, 1
	n6 := CreateTestContextFile("context", "6", "dd6").PushDep(n4.DataDepDig).PushDep(n5.DataDepDig)
	n6.Snapshot = true
	w6, err6 := db.StoreContextFile(n6)
	if err6 != nil || !(w6.Depth == 0 && w6.Head && !w6.Leaf && !w6.CoveredBySnapshot) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 1 && !w5.Head && !w5.Leaf && w5.CoveredBySnapshot) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 1 && !w4.Head && !w4.Leaf && w4.CoveredBySnapshot) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 2 && !w3.Head && w3.Leaf && w3.CoveredBySnapshot) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 2 && !w2.Head && !w2.Leaf && w2.CoveredBySnapshot) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 3 && !w1.Head && w1.Leaf && w1.CoveredBySnapshot) {
		t.Error()
	}

	// 7: 4, 5, 3, 2, 1
	n7 := CreateTestContextFile("context", "7", "dd7").PushDep(n6.DataDepDig)
	w7, err7 := db.StoreContextFile(n7)
	if err7 != nil || !(w7.Depth == 0 && w7.Head && !w7.Leaf && !w7.CoveredBySnapshot) {
		t.Error()
	}
	w6 = db.GetContextFileByDepDataDig(n6.DataDepDig)
	if !(w6.Depth == 1 && !w6.Head && !w6.Leaf && !w6.CoveredBySnapshot) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 2 && !w5.Head && !w5.Leaf && w5.CoveredBySnapshot) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 2 && !w4.Head && !w4.Leaf && w4.CoveredBySnapshot) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 3 && !w3.Head && w3.Leaf && w3.CoveredBySnapshot) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 3 && !w2.Head && !w2.Leaf && w2.CoveredBySnapshot) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 4 && !w1.Head && w1.Leaf && w1.CoveredBySnapshot) {
		t.Error()
	}

	// 8: 4, 3
	n8 := CreateTestContextFile("context", "8", "dd8").PushDep(n5.DataDepDig)
	w8, err8 := db.StoreContextFile(n8)
	if err8 != nil || !(w8.Depth == 0 && w8.Head && !w8.Leaf && !w8.CoveredBySnapshot) {
		t.Error()
	}
	w7 = db.GetContextFileByDepDataDig(n7.DataDepDig)
	if !(w7.Depth == 0 && w7.Head && !w7.Leaf && !w7.CoveredBySnapshot) {
		t.Error()
	}
	w6 = db.GetContextFileByDepDataDig(n6.DataDepDig)
	if !(w6.Depth == 1 && !w6.Head && !w6.Leaf && !w6.CoveredBySnapshot) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 2 && !w5.Head && !w5.Leaf && !w5.CoveredBySnapshot) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 2 && !w4.Head && !w4.Leaf && w4.CoveredBySnapshot) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 3 && !w3.Head && w3.Leaf && w3.CoveredBySnapshot) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 3 && !w2.Head && !w2.Leaf && !w2.CoveredBySnapshot) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 4 && !w1.Head && w1.Leaf && !w1.CoveredBySnapshot) {
		t.Error()
	}

	n9 := CreateTestContextFile("context", "9", "dd9").PushDep(n4.DataDepDig)
	w9, err9 := db.StoreContextFile(n9)
	if err9 != nil || !(w9.Depth == 0 && w9.Head && !w9.Leaf && !w9.CoveredBySnapshot) {
		t.Error()
	}
	w8 = db.GetContextFileByDepDataDig(n8.DataDepDig)
	if !(w8.Depth == 0 && w8.Head && !w8.Leaf && !w8.CoveredBySnapshot) {
		t.Error()
	}
	w7 = db.GetContextFileByDepDataDig(n7.DataDepDig)
	if !(w7.Depth == 0 && w7.Head && !w7.Leaf && !w7.CoveredBySnapshot) {
		t.Error()
	}
	w6 = db.GetContextFileByDepDataDig(n6.DataDepDig)
	if !(w6.Depth == 1 && !w6.Head && !w6.Leaf && !w6.CoveredBySnapshot) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 2 && !w5.Head && !w5.Leaf && !w5.CoveredBySnapshot) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 2 && !w4.Head && !w4.Leaf && !w4.CoveredBySnapshot) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 3 && !w3.Head && w3.Leaf && !w3.CoveredBySnapshot) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 3 && !w2.Head && !w2.Leaf && !w2.CoveredBySnapshot) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 4 && !w1.Head && w1.Leaf && !w1.CoveredBySnapshot) {
		t.Error()
	}

	n10 := CreateTestContextFile("context", "10", "dd10").PushDep(n9.DataDepDig)
	w10, err10 := db.StoreContextFile(n10)
	if err10 != nil || !(w10.Depth == 0 && w10.Head && !w10.Leaf) {
		t.Error()
	}
	w9 = db.GetContextFileByDepDataDig(n9.DataDepDig)
	if !(w9.Depth == 1 && !w9.Head && !w9.Leaf) {
		t.Error()
	}
	w8 = db.GetContextFileByDepDataDig(n8.DataDepDig)
	if !(w8.Depth == 0 && w8.Head && !w8.Leaf) {
		t.Error()
	}
	w7 = db.GetContextFileByDepDataDig(n7.DataDepDig)
	if !(w7.Depth == 0 && w7.Head && !w7.Leaf) {
		t.Error()
	}
	w6 = db.GetContextFileByDepDataDig(n6.DataDepDig)
	if !(w6.Depth == 1 && !w6.Head && !w6.Leaf) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 2 && !w5.Head && !w5.Leaf) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 2 && !w4.Head && !w4.Leaf) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 3 && !w3.Head && w3.Leaf) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 3 && !w2.Head && !w2.Leaf) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 4 && !w1.Head && w1.Leaf) {
		t.Error()
	}

	n11 := CreateTestContextFile("context", "11", "dd11").PushDep(n8.DataDepDig)
	w11, err11 := db.StoreContextFile(n11)
	if err11 != nil || !(w11.Depth == 0 && w11.Head && !w11.Leaf) {
		t.Error()
	}
	w10 = db.GetContextFileByDepDataDig(n10.DataDepDig)
	if !(w10.Depth == 0 && w10.Head && !w10.Leaf) {
		t.Error()
	}
	w9 = db.GetContextFileByDepDataDig(n9.DataDepDig)
	if !(w9.Depth == 1 && !w9.Head && !w9.Leaf) {
		t.Error()
	}
	w8 = db.GetContextFileByDepDataDig(n8.DataDepDig)
	if !(w8.Depth == 1 && !w8.Head && !w8.Leaf) {
		t.Error()
	}
	w7 = db.GetContextFileByDepDataDig(n7.DataDepDig)
	if !(w7.Depth == 0 && w7.Head && !w7.Leaf) {
		t.Error()
	}
	w6 = db.GetContextFileByDepDataDig(n6.DataDepDig)
	if !(w6.Depth == 1 && !w6.Head && !w6.Leaf) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 2 && !w5.Head && !w5.Leaf) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 2 && !w4.Head && !w4.Leaf) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 3 && !w3.Head && w3.Leaf) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 3 && !w2.Head && !w2.Leaf) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 4 && !w1.Head && w1.Leaf) {
		t.Error()
	}

	n12 := CreateTestContextFile("context", "12", "dd12").PushDep(n6.DataDepDig)
	w12, err12 := db.StoreContextFile(n12)
	if err12 != nil || !(w12.Depth == 0 && w12.Head && !w12.Leaf) {
		t.Error()
	}
	w11 = db.GetContextFileByDepDataDig(n11.DataDepDig)
	if !(w11.Depth == 0 && w11.Head && !w11.Leaf) {
		t.Error()
	}
	w10 = db.GetContextFileByDepDataDig(n10.DataDepDig)
	if !(w10.Depth == 0 && w10.Head && !w10.Leaf) {
		t.Error()
	}
	w9 = db.GetContextFileByDepDataDig(n9.DataDepDig)
	if !(w9.Depth == 1 && !w9.Head && !w9.Leaf) {
		t.Error()
	}
	w8 = db.GetContextFileByDepDataDig(n8.DataDepDig)
	if !(w8.Depth == 1 && !w8.Head && !w8.Leaf) {
		t.Error()
	}
	w7 = db.GetContextFileByDepDataDig(n7.DataDepDig)
	if !(w7.Depth == 0 && w7.Head && !w7.Leaf) {
		t.Error()
	}
	w6 = db.GetContextFileByDepDataDig(n6.DataDepDig)
	if !(w6.Depth == 1 && !w6.Head && !w6.Leaf) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 2 && !w5.Head && !w5.Leaf) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 2 && !w4.Head && !w4.Leaf) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 3 && !w3.Head && w3.Leaf) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 3 && !w2.Head && !w2.Leaf) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 4 && !w1.Head && w1.Leaf) {
		t.Error()
	}

	n13 := CreateTestContextFile("context", "13", "dd13").PushDep(n2.DataDepDig)
	w13, err13 := db.StoreContextFile(n13)
	if err13 != nil || !(w13.Depth == 0 && w13.Head && !w13.Leaf) {
		t.Error()
	}
	w12 = db.GetContextFileByDepDataDig(n12.DataDepDig)
	if !(w12.Depth == 0 && w12.Head && !w12.Leaf) {
		t.Error()
	}
	w11 = db.GetContextFileByDepDataDig(n11.DataDepDig)
	if !(w11.Depth == 0 && w11.Head && !w11.Leaf) {
		t.Error()
	}
	w10 = db.GetContextFileByDepDataDig(n10.DataDepDig)
	if !(w10.Depth == 0 && w10.Head && !w10.Leaf) {
		t.Error()
	}
	w9 = db.GetContextFileByDepDataDig(n9.DataDepDig)
	if !(w9.Depth == 1 && !w9.Head && !w9.Leaf) {
		t.Error()
	}
	w8 = db.GetContextFileByDepDataDig(n8.DataDepDig)
	if !(w8.Depth == 1 && !w8.Head && !w8.Leaf) {
		t.Error()
	}
	w7 = db.GetContextFileByDepDataDig(n7.DataDepDig)
	if !(w7.Depth == 0 && w7.Head && !w7.Leaf) {
		t.Error()
	}
	w6 = db.GetContextFileByDepDataDig(n6.DataDepDig)
	if !(w6.Depth == 1 && !w6.Head && !w6.Leaf) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 2 && !w5.Head && !w5.Leaf) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 2 && !w4.Head && !w4.Leaf) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 3 && !w3.Head && w3.Leaf) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 3 && !w2.Head && !w2.Leaf) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 4 && !w1.Head && w1.Leaf) {
		t.Error()
	}

	n14 := CreateTestContextFile("context", "14", "dd14")
	n14.Snapshot = true
	w14, err14 := db.StoreContextFile(n14)
	if err14 != nil || !(w14.Depth == 0 && w14.Head && w14.Leaf) {
		t.Error()
	}
	w13 = db.GetContextFileByDepDataDig(n13.DataDepDig)
	if !(w13.Depth == 0 && w13.Head && !w13.Leaf) {
		t.Error()
	}
	w12 = db.GetContextFileByDepDataDig(n12.DataDepDig)
	if !(w12.Depth == 0 && w12.Head && !w12.Leaf) {
		t.Error()
	}
	w11 = db.GetContextFileByDepDataDig(n11.DataDepDig)
	if !(w11.Depth == 0 && w11.Head && !w11.Leaf) {
		t.Error()
	}
	w10 = db.GetContextFileByDepDataDig(n10.DataDepDig)
	if !(w10.Depth == 0 && w10.Head && !w10.Leaf) {
		t.Error()
	}
	w9 = db.GetContextFileByDepDataDig(n9.DataDepDig)
	if !(w9.Depth == 1 && !w9.Head && !w9.Leaf) {
		t.Error()
	}
	w8 = db.GetContextFileByDepDataDig(n8.DataDepDig)
	if !(w8.Depth == 1 && !w8.Head && !w8.Leaf) {
		t.Error()
	}
	w7 = db.GetContextFileByDepDataDig(n7.DataDepDig)
	if !(w7.Depth == 0 && w7.Head && !w7.Leaf) {
		t.Error()
	}
	w6 = db.GetContextFileByDepDataDig(n6.DataDepDig)
	if !(w6.Depth == 1 && !w6.Head && !w6.Leaf) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 2 && !w5.Head && !w5.Leaf) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 2 && !w4.Head && !w4.Leaf) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 3 && !w3.Head && w3.Leaf) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 3 && !w2.Head && !w2.Leaf) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 4 && !w1.Head && w1.Leaf) {
		t.Error()
	}

	n15 := CreateTestContextFile("context", "15", "dd15").PushDep(n7.DataDepDig)
	n15.Snapshot = true
	w15, err15 := db.StoreContextFile(n15)
	if err15 != nil || !(w15.Depth == 0 && w15.Head && !w15.Leaf && !w15.CoveredBySnapshot) {
		t.Error()
	}
	w14 = db.GetContextFileByDepDataDig(n14.DataDepDig)
	if !(w14.Depth == 0 && w14.Head && w14.Leaf && !w14.CoveredBySnapshot) {
		t.Error()
	}
	w13 = db.GetContextFileByDepDataDig(n13.DataDepDig)
	if !(w13.Depth == 0 && w13.Head && !w13.Leaf && !w13.CoveredBySnapshot) {
		t.Error()
	}
	w12 = db.GetContextFileByDepDataDig(n12.DataDepDig)
	if !(w12.Depth == 0 && w12.Head && !w12.Leaf && !w12.CoveredBySnapshot) {
		t.Error()
	}
	w11 = db.GetContextFileByDepDataDig(n11.DataDepDig)
	if !(w11.Depth == 0 && w11.Head && !w11.Leaf && !w11.CoveredBySnapshot) {
		t.Error()
	}
	w10 = db.GetContextFileByDepDataDig(n10.DataDepDig)
	if !(w10.Depth == 0 && w10.Head && !w10.Leaf && !w10.CoveredBySnapshot) {
		t.Error()
	}
	w9 = db.GetContextFileByDepDataDig(n9.DataDepDig)
	if !(w9.Depth == 1 && !w9.Head && !w9.Leaf && !w9.CoveredBySnapshot) {
		t.Error()
	}
	w8 = db.GetContextFileByDepDataDig(n8.DataDepDig)
	if !(w8.Depth == 1 && !w8.Head && !w8.Leaf && !w8.CoveredBySnapshot) {
		t.Error()
	}
	w7 = db.GetContextFileByDepDataDig(n7.DataDepDig)
	if !(w7.Depth == 1 && !w7.Head && !w7.Leaf && w7.CoveredBySnapshot) {
		t.Error()
	}
	w6 = db.GetContextFileByDepDataDig(n6.DataDepDig)
	if !(w6.Depth == 2 && !w6.Head && !w6.Leaf && !w6.CoveredBySnapshot) {
		t.Error()
	}
	w5 = db.GetContextFileByDepDataDig(n5.DataDepDig)
	if !(w5.Depth == 3 && !w5.Head && !w5.Leaf && !w5.CoveredBySnapshot) {
		t.Error()
	}
	w4 = db.GetContextFileByDepDataDig(n4.DataDepDig)
	if !(w4.Depth == 3 && !w4.Head && !w4.Leaf && !w4.CoveredBySnapshot) {
		t.Error()
	}
	w3 = db.GetContextFileByDepDataDig(n3.DataDepDig)
	if !(w3.Depth == 4 && !w3.Head && w3.Leaf && !w3.CoveredBySnapshot) {
		t.Error()
	}
	w2 = db.GetContextFileByDepDataDig(n2.DataDepDig)
	if !(w2.Depth == 4 && !w2.Head && !w2.Leaf && !w2.CoveredBySnapshot) {
		t.Error()
	}
	w1 = db.GetContextFileByDepDataDig(n1.DataDepDig)
	if !(w1.Depth == 5 && !w1.Head && w1.Leaf && !w1.CoveredBySnapshot) {
		t.Error()
	}
}
