// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
)

// A DebugTree captures a hierarchical debug trace. It's useful for
// visualizing the execution of recursive functions.
type DebugTree struct {
	cur   *debugTreeNode
	roots []*debugTreeNode

	nextEdge string
}

type debugTreeNode struct {
	label    string
	parent   *debugTreeNode
	edges    []string
	children []*debugTreeNode
}

func (t *DebugTree) Push(label string) {
	node := &debugTreeNode{label: label, parent: t.cur}
	if t.cur == nil {
		t.roots = append(t.roots, node)
	} else {
		t.cur.edges = append(t.cur.edges, t.nextEdge)
		t.cur.children = append(t.cur.children, node)
	}
	t.cur = node
	t.nextEdge = ""
}

func (t *DebugTree) Pushf(format string, args ...interface{}) {
	t.Push(fmt.Sprintf(format, args...))
}

func (t *DebugTree) Append(label string) {
	t.cur.label += label
}

func (t *DebugTree) Appendf(format string, args ...interface{}) {
	t.Append(fmt.Sprintf(format, args...))
}

func (t *DebugTree) Pop() {
	if t.cur == nil {
		panic("unbalanced Push/Pop")
	}
	t.cur = t.cur.parent
	t.nextEdge = ""
}

func (t *DebugTree) Leaf(label string) {
	t.Push(label)
	t.Pop()
}

func (t *DebugTree) Leaff(format string, args ...interface{}) {
	t.Leaf(fmt.Sprintf(format, args...))
}

func (t *DebugTree) SetEdge(label string) {
	t.nextEdge = label
}

func (t *DebugTree) WriteToDot(w io.Writer) {
	id := func(n *debugTreeNode) string {
		return fmt.Sprintf("n%p", n)
	}

	var rec func(n *debugTreeNode)
	rec = func(n *debugTreeNode) {
		nid := id(n)
		fmt.Fprintf(w, "%s [label=%q];\n", nid, n.label)
		for i, child := range n.children {
			fmt.Fprintf(w, "%s -> %s", nid, id(child))
			if n.edges[i] != "" {
				fmt.Fprintf(w, " [label=%q]", n.edges[i])
			}
			fmt.Fprint(w, ";\n")
			rec(child)
		}
	}

	fmt.Fprint(w, "digraph debug {\n")
	for _, root := range t.roots {
		rec(root)
	}
	fmt.Fprint(w, "}\n")
}

type IndentWriter struct {
	W      io.Writer
	Indent []byte
	inLine bool
}

func (w *IndentWriter) Write(p []byte) (n int, err error) {
	total := 0
	for len(p) > 0 {
		if !w.inLine {
			_, err := w.W.Write(w.Indent)
			if err != nil {
				return total, err
			}
			w.inLine = true
		}

		next := bytes.IndexByte(p, '\n')
		if next < 0 {
			n, err := w.W.Write(p)
			total += n
			return total, err
		}
		line, rest := p[:next+1], p[next+1:]
		n, err := w.W.Write(line)
		total += n
		if n < len(line) || err != nil {
			return total, err
		}
		w.inLine = false
		p = rest
	}
	return total, nil
}
