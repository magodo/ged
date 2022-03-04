package pkg1

func F1() {}

type T1 struct {
	T1F1 int
	T2
}

type T2 struct {
	T2F1 int
}

func (t1 T1) F1()  {}
func (t1 *T1) F2() {}
