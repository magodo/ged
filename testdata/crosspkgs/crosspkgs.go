package crosspkgs

import (
	"uut/pkg1"
	"uut/pkg2"
)

func crosspkgs() {
	t1 := pkg1.T1{}
	t2 := pkg2.T1{}
	t1.F1()
	t2.F1()
	t2.F2()
}
