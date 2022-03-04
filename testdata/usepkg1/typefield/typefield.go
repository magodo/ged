package typefield

import (
	"uut/pkg1"
)

func pkgfield() {
	t2 := pkg1.T2{}
	t1 := pkg1.T1{T2: t2}
	_ = t1.T1F1
	_ = t1.T2F1
	_ = t1.T2
}
