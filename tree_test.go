package gomt

import (
	"testing"
)

func compare(t *testing.T, a *node, b *node) bool {
	if a.id != b.id || a.Mod != b.Mod || a.Version != b.Version || len(a.Deps) != len(b.Deps) {
		t.Log("unexpect",
			a.id, a.Mod, a.Version, len(a.Deps),
			b.id, b.Mod, b.Version, len(b.Deps))
		return false
	}
	for i := range a.Deps {
		if !compare(t, a.Deps[i], b.Deps[i]) {
			return false
		}
	}
	return true
}

func TestParseGoModGraph(t *testing.T) {
	graph := `a b@v1.0.0
b@v1.0.0 c@v0.0.8
c@v0.0.8 golang.org/x/crypto@v0.0.0-20181203042331-505ab145d0a9
c@v0.0.8 golang.org/x/sys@v0.0.0-20181205085412-a5c9d58dba9a
a f@v2.0.0
f@v2.0.0 g@v0.0.3
f@v2.0.0 h@v0.0.4`
	root, err := parseGoModGraph(graph)
	if err != nil {
		t.Error(err)
	}
	a := newNode("a")
	b := newNode("b@v1.0.0")
	c := newNode("c@v0.0.8")
	d := newNode("golang.org/x/crypto@v0.0.0-20181203042331-505ab145d0a9")
	e := newNode("golang.org/x/sys@v0.0.0-20181205085412-a5c9d58dba9a")
	f := newNode("f@v2.0.0")
	g := newNode("g@v0.0.3")
	h := newNode("h@v0.0.4")
	a.Deps = []*node{b, f}
	b.Deps = []*node{c}
	c.Deps = []*node{d, e}
	f.Deps = []*node{g, h}

	if !compare(t, root, a) {
		t.Error("unexpected parse graph")
	}
}

func TestFilterGraph(t *testing.T) {
	graph := `a b@v1.0.0
b@v1.0.0 c@v0.0.8
c@v0.0.8 golang.org/x/crypto@v0.0.0-20181203042331-505ab145d0a9
c@v0.0.8 golang.org/x/sys@v0.0.0-20181205085412-a5c9d58dba9a
a f@v2.0.0
f@v2.0.0 g@v0.0.3
f@v2.0.0 h@v0.0.4`
	{
		root, err := parseGoModGraph(graph)
		if err != nil {
			t.Error(err)
		}
		root = filterGraph(root, options{maxLevel: 1})
		a := newNode("a")
		b := newNode("b@v1.0.0")
		f := newNode("f@v2.0.0")
		a.Deps = []*node{b, f}
		if !compare(t, root, a) {
			t.Error("unexpected parse graph")
		}
	}
	{
		root, err := parseGoModGraph(graph)
		if err != nil {
			t.Error(err)
		}
		root = filterGraph(root, options{target: "c"})
		a := newNode("a")
		b := newNode("b@v1.0.0")
		c := newNode("c@v0.0.8")
		a.Deps = []*node{b}
		b.Deps = []*node{c}
		if !compare(t, root, a) {
			t.Error("unexpected parse graph")
		}
	}
}

func TestTree(t *testing.T) {
	graph := `a b
b c
c d
c e
a f
f g
f h`
	root, err := parseGoModGraph(graph)
	if err != nil {
		t.Error(err)
	}
	{
		result := tree(root, options{})
		if result != `a
├── b
│   └── c
│       ├── d
│       └── e
└── f
    ├── g
    └── h` {
			t.Log(result)
			t.Error("unexpected tree, no target")
		}
	}
}
