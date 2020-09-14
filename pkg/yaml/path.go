package yaml

import (
	"strconv"
	"strings"
)

type Path struct {
	isInt bool
	i int
	s string
	parent *Path
}

func (p *Path) AddInt(val int) *Path {
	newPath := Path{
		isInt: true,
		i: val,
		s: "",
		parent: p,
	}
	return &newPath
}

func (p *Path) AddString(val string) *Path {
	newPath := Path{
		isInt: false,
		i: 0,
		s: val,
		parent: p,
	}
	return &newPath
}

func (p *Path) String() string {
	if p == nil {
		return ""
	}
	var out []string
	for entry := p; entry.parent != nil; entry = entry.parent {
		if entry.isInt {
			out = append([]string{strconv.Itoa(entry.i)}, out...)
		} else {
			out = append([]string{strconv.Quote(entry.s)}, out...)
		}
	}
	return strings.Join(out, ".")
}
