package main

import (
	"fmt"
	"go/token"
	"sort"
)

type Match struct {
	pkg    string
	ident  string
	method string
	field  string
}

func (m Match) String() string {
	msg := fmt.Sprintf("%s %s", m.pkg, m.ident)
	if m.method != "" {
		return msg + "." + m.method + "()"
	}
	if m.field != "" {
		return msg + "." + m.field
	}
	return msg
}

type MatchSlice []Match

func (ms MatchSlice) Len() int {
	return len(ms)
}

func (ms MatchSlice) Less(i, j int) bool {
	if ms[i].pkg != ms[j].pkg {
		return ms[i].pkg < ms[j].pkg
	}
	if ms[i].ident != ms[j].ident {
		return ms[i].ident < ms[j].ident
	}
	if ms[i].method != ms[j].method {
		return ms[i].method < ms[j].method
	}
	return ms[i].field < ms[j].field
}

func (ms MatchSlice) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

type Matches map[Match]Positions

func (ms Matches) Add(m Match, pos token.Position) {
	poses, ok := ms[m]
	if !ok {
		poses = Positions{}
		ms[m] = poses
	}
	poses[pos] = true
}

func (ms Matches) Merge(oms Matches) {
	for m, poses := range oms {
		if _, ok := ms[m]; !ok {
			ms[m] = poses
			continue
		}
		for pos := range poses {
			ms[m][pos] = true
		}
	}
}

func (ms Matches) String() string {
	var matchSlice MatchSlice
	for m := range ms {
		matchSlice = append(matchSlice, m)
	}
	sort.Sort(matchSlice)

	var out string
	for _, m := range matchSlice {
		out += m.String() + ":\n"
		for _, pos := range ms[m].Positions() {
			out += "\t" + pos + "\n"
		}
	}
	return out
}

type Positions map[token.Position]bool

type PositionSlice []token.Position

func (poses PositionSlice) Len() int {
	return len(poses)
}

func (poses PositionSlice) Less(i, j int) bool {
	pos1, pos2 := poses[i], poses[j]
	if pos1.Filename != pos2.Filename {
		return pos1.Filename < pos2.Filename
	}
	if pos1.Line != pos2.Line {
		return pos1.Line < pos2.Line
	}
	if pos1.Column != pos2.Column {
		return pos1.Column < pos2.Column
	}
	return pos1.Offset < pos2.Offset
}

func (poses PositionSlice) Swap(i, j int) {
	poses[i], poses[j] = poses[j], poses[i]
}

func (poses Positions) Positions() []string {
	ps := make(PositionSlice, 0, len(poses))
	out := make([]string, 0, len(poses))
	for pos := range poses {
		ps = append(ps, pos)
	}
	sort.Sort(ps)
	for _, pos := range ps {
		out = append(out, pos.String())
	}
	return out
}
