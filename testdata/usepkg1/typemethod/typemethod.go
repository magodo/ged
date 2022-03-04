package typemethod

import (
	"uut/pkg1"
)

func typemethod() {
	t1 := pkg1.T1{}
	t1.F1()
	t1.F2()

	pt1 := &pkg1.T1{}
	pt1.F1()
	pt1.F2()
}
