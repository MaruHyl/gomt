package gomt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func Tree(options ...Option) (string, error) {
	opts, err := buildOpts(options...)
	if err != nil {
		return "", err
	}
	// get go mod graph
	graph, err := runGoModGraph()
	if err != nil {
		return "", err
	}
	// parse go mod graph
	root, err := parseGoModGraph(graph)
	if err != nil {
		return "", err
	}
	if root == nil {
		return "", nil
	}
	// filter go mod graph
	root = filterGraph(root, opts)
	// draw graph
	if opts.json {
		b, err := json.MarshalIndent(root, "", " ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return tree(root, opts), nil
}

func tree(root *node, opts options) string {
	const prefixClose = "└── "
	const prefixOpen = "├── "
	const close = "    "
	const open = "│   "

	lines := make([]string, 0)
	var visit func(n *node, flags []bool)
	visit = func(n *node, flags []bool) {
		// build prefix
		prefix := ""
		for i, flag := range flags {
			isPrefix := i == len(flags)-1
			if flag {
				if isPrefix {
					prefix += prefixOpen
				} else {
					prefix += open
				}
			} else {
				if isPrefix {
					prefix += prefixClose
				} else {
					prefix += close
				}
			}
		}
		id := n.getID()
		if opts.target != "" && n.Mod == opts.target {
			id = color.RedString(id)
		}
		lines = append(lines, prefix+id)
		// traversal adj
		for i, adj := range n.Deps {
			isOpen := true
			if i == len(n.Deps)-1 {
				isOpen = false
			}
			flags = append(flags, isOpen)
			visit(adj, flags)
			flags = flags[:len(flags)-1]
		}
	}
	visit(root, nil)
	return strings.Join(lines, "\n")
}

func filterGraph(root *node, opts options) *node {
	nodeMap := make(map[string]*node)
	var visit func(n *node, level int) bool
	visit = func(n *node, level int) bool {
		if opts.maxLevel > 0 && level > opts.maxLevel {
			return false
		}
		join := opts.target == "" || n.Mod == opts.target
		for _, adj := range n.Deps {
			if visit(adj, level+1) {
				join = true
				putEdge(nodeMap, n.getID(), adj.getID())
			}
		}
		return join
	}
	if visit(root, 0) {
		return getOrCreateNode(nodeMap, root.getID())
	}
	return nil
}

type node struct {
	id      string
	Mod     string
	Version string
	Deps    []*node `json:",omitempty"`
}

func (n node) getID() string {
	return n.id
}

func newNode(id string) *node {
	mod, version := splitID(id)
	return &node{
		id:      id,
		Mod:     mod,
		Version: version,
	}
}

func putEdge(nodeMap map[string]*node, fromID string, toID string) {
	fromNode := getOrCreateNode(nodeMap, fromID)
	toNode := getOrCreateNode(nodeMap, toID)
	fromNode.Deps = append(fromNode.Deps, toNode)
}

func getOrCreateNode(nodeMap map[string]*node, id string) *node {
	n, ok := nodeMap[id]
	if !ok {
		n = newNode(id)
		nodeMap[id] = n
	}
	return n
}

func splitID(id string) (mod string, version string) {
	s := strings.SplitN(id, "@", 2)
	mod = s[0]
	if len(s) > 1 {
		version = s[1]
	}
	return
}

func combineID(mod string, version string) string {
	if version == "" {
		return mod
	}
	return mod + "@" + version
}

func parseGoModGraph(result string) (root *node, err error) {
	if result == "" {
		return
	}
	lines := strings.Split(result, "\n")
	nodeMap := make(map[string]*node)
	findRoot := false
	for _, line := range lines {
		if line == "" {
			continue
		}
		// skip go findings
		if strings.HasPrefix(line, "go:") {
			continue
		}
		edge := strings.Split(line, " ")
		if len(edge) != 2 {
			return nil, fmt.Errorf("unexpected parse error: %s", line)
		}
		putEdge(nodeMap, edge[0], edge[1])
		if !findRoot {
			root = getOrCreateNode(nodeMap, edge[0])
			findRoot = true
		}
	}
	// TODO already sorted by go mod graph
	//for _, n := range nodeMap {
	//	sort.Slice(n.Deps, func(i, j int) bool {
	//		return n.Deps[i].getID() < n.Deps[j].getID()
	//	})
	//}
	return
}

func runGoModGraph() (string, error) {
	cmd := exec.Command("go", "mod", "graph")
	cmd.Env = getCmdEnv()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%v: %s", err, stderr.String())
	}
	return stdout.String(), err
}

func getCmdEnv() []string {
	goPath := os.Getenv("GOPATH")
	path := fmt.Sprintf("PATH=%s", os.Getenv("PATH"))
	home := fmt.Sprintf("HOME=%s", os.Getenv("HOME"))
	cgo := "CGO_ENABLED=0"
	goMod := "GO111MODULE=on"
	goPathEnv := fmt.Sprintf("GOPATH=%s", goPath)
	goCache := fmt.Sprintf("GOCACHE=%s", filepath.Join(goPath, "cache"))
	return []string{
		path,
		home,
		cgo,
		goMod,
		goPathEnv,
		goCache,
	}
}
