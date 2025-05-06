package myj

import (
	"fmt"
	"io"
	"log"

	"github.com/monopole/gojira/internal/utils"
)

const FlagDoIt = "go"

// Graph is a directed graph holding nodes and edges to ease
// finding and fixing date anomalies.
type Graph struct {
	nodes map[MyKey]*Node
	edges map[Edge]bool
}

type Node struct {
	// issue is meant to be immutable here
	issue *ResponseIssue

	// dependsOn and isDependedOnBy hold the contents of the `edge` map in a
	// form that's more convenient for traversal; see loadEdgesIntoNodes.
	dependsOn      []*Node
	isDependedOnBy []*Node

	// mutable fields follow.  The date fields are used to store proposed
	// new dates to use in date repair code.
	dateStart  utils.Date
	dateEnd    utils.Date
	visitCount int
}

type Edge struct {
	// In issue terms, a "parent" is a dependency -
	// something that must be done before the child can be started.
	parent MyKey
	// A "child" is a dependent; it cannot start until the parent is done.
	child MyKey
}

// MakeGraph returns an instance of Graph to wrap a set of nodes and edges
// in convenience methods.  It's assumed that the arguments (nodes and edges)
// already make up a proper directed graph.
func MakeGraph(nodes map[MyKey]*Node, edges map[Edge]bool) *Graph {
	g := &Graph{nodes: nodes, edges: edges}
	g.loadEdgesIntoNodes()
	return g
}

func (g *Graph) Nodes() map[MyKey]*Node {
	return g.nodes
}

func MakeNode(epic *ResponseIssue) *Node {
	return &Node{
		issue:     epic,
		dateStart: epic.DateStart(),
		dateEnd:   epic.DateEnd(),
	}
}

func (n *Node) seemsDone() bool {
	return n.issue.Status() == IssueStatusClosed ||
		n.issue.Status() == IssueStatusDone ||
		n.issue.Status() == IssueStatusClosedWoAction
}

func (n *Node) writeDiGraphNode(w io.Writer) {
	_, _ = fmt.Fprintf(
		w,
		"  %q [label=\"%s\" style=filled fillcolor=%s];\n",
		n.issue.MyKey, n.digraphLabel(),
		StatusColor(n.issue.Status(), ColorKindDot))
}

func (n *Node) digraphLabel() string {
	return fmt.Sprintf(
		"%s\n%s %s\n%s",
		n.issue.MyKey,
		func() string {
			dr, err := utils.MakeDayRangeGentle(n.dateStart, n.dateEnd)
			if err != nil {
				return "(" + utils.Ellipsis(err.Error(), 20) + ")"
			}
			return dr.PrettyRange()
		}(),
		n.issue.AssigneeLdap(),
		utils.ShortLines(n.issue.MySummary()))
}

// loadEdgesIntoNodes just makes it easier to compute in-degree/out-degree and
// traverse edges, rather than repeatedly scanning all edges looking for node
// matches.
func (g *Graph) loadEdgesIntoNodes() {
	for edge := range g.edges {
		g.nodes[edge.child].dependsOn = append(
			g.nodes[edge.child].dependsOn, g.nodes[edge.parent])
		g.nodes[edge.parent].isDependedOnBy = append(
			g.nodes[edge.parent].isDependedOnBy, g.nodes[edge.child])
	}
}

func (g *Graph) WriteDigraph(w io.Writer, flip bool) {
	_, _ = fmt.Fprintln(w, "digraph dependencies {")
	_, _ = fmt.Fprintf(w, "  rankdir=%s;\n",
		func() string {
			if flip {
				return "BT"
			}
			return "TB"
		}())
	_, _ = fmt.Fprintln(w, "  node [shape=ellipse];")
	for _, node := range g.nodes {
		node.writeDiGraphNode(w)
	}
	for edge := range g.edges {
		_, _ = fmt.Fprintf(w, "%q -> %q;\n", edge.parent, edge.child)
	}
	_, _ = fmt.Fprintln(w, "}")
}

func (g *Graph) ReportMisOrdering(w io.Writer) {
	for edge := range g.edges {
		start := g.nodes[edge.child].dateStart
		end := g.nodes[edge.parent].dateEnd
		if !start.After(end) {
			_, _ = fmt.Fprintf(w,
				"%10s depends on %10s, but%4d starts on %s, %3d days before%4d ends on %s.\n",
				edge.child, edge.parent,
				edge.child.Num, start.Brief(), start.DayCount(end),
				edge.parent.Num, end.Brief())
		}
	}
}

func (g *Graph) ReportWeekends(w io.Writer) {
	for _, n := range g.nodes {
		if n.dateStart.IsWeekend() {
			_, _ = fmt.Fprintf(w, "%12s starts on a %s (%s), pushing to Mon.\n",
				n.issue.MyKey, n.dateStart.Weekday(), n.dateStart.Brief())
			n.dateStart = n.dateStart.SlideOverWeekend()
		}
		if n.dateEnd.IsWeekend() {
			_, _ = fmt.Fprintf(w, "%12s ends on a %s (%s), pulling to Fri.\n",
				n.issue.MyKey, n.dateEnd.Weekday(), n.dateEnd.Brief())
			n.dateEnd = n.dateEnd.SlideBeforeWeekend()
		}
	}
}

func (g *Graph) ScanAndReportNodes(w io.Writer) {
	for key := range g.nodes {
		node := g.nodes[key]
		dr, dateErr := utils.MakeDayRangeGentle(node.dateStart, node.dateEnd)
		if dateErr != nil {
			node.dateStart = dr.Start()
			node.dateEnd = dr.End()
		}
		_, _ = fmt.Fprintf(
			w,
			"%12s has %3d parents, %3d children %s %s\n",
			key.String(),
			len(node.dependsOn),
			len(node.isDependedOnBy),
			func() string {
				if len(node.dependsOn) == 0 &&
					len(node.isDependedOnBy) == 0 {
					return "Isolated epic!"
				}
				if len(node.dependsOn) == 0 {
					return "ROOT"
				}
				if len(node.isDependedOnBy) == 0 {
					return "LEAF"
				}
				return ""
			}(),
			func() string {
				if dateErr != nil {
					return dateErr.Error()
				}
				return ""
			}(),
		)
	}
}

func (g *Graph) resetVisits() {
	for _, node := range g.nodes {
		node.visitCount = 0
	}
}

// MaybeShiftDependentsLater might push dependent ("child") epics out in time
// to start after their dependencies ("parents") end.
func (g *Graph) MaybeShiftDependentsLater() {
	g.resetVisits()
	for _, node := range g.nodes {
		if len(node.dependsOn) == 0 {
			// This node depends on nothing; it's a root, and
			// is the entry point into the digraph.
			node.MaybeShiftDependentsLater()
		}
	}
}

// MaybeShiftEarlier tries to tighten up the schedule without
// violating dependencies.
func (g *Graph) MaybeShiftEarlier() {
	g.resetVisits()
	for _, node := range g.nodes {
		if len(node.isDependedOnBy) == 0 {
			// This node is a leaf, presumably a project endpoint as
			// nothing depends on it.
			node.MaybeShiftEarlier()
		}
	}
}

// MaybeShiftDependentsLater wants a graph in which no child starts
// before the parent ends.
// I.e. it shifts dependents (children) later if they start before
// their dependency (parent) completes.
func (n *Node) MaybeShiftDependentsLater() {
	n.cycleCheck()
	for _, child := range n.isDependedOnBy {
		if child.seemsDone() {
			continue
		}
		if !child.dateStart.After(n.dateEnd) {
			// This is a problem.

			// A given child may have more than one parent in a digraph,
			// e.g. where several epics block, say, the final epic.
			// Each of these 'parent' epics might have a different completion
			// date. We visit the child once for each parent, moving
			// its start date further out to begin after whichever parent
			// ends the latest.

			saveDayCount := child.dateStart.DayCount(child.dateEnd)

			// Move child start to one day after parent end.
			child.dateStart = n.dateEnd.AddDays(1).SlideOverWeekend()

			// Move child end to roughly establish the same duration
			// as before, modulo not ending on a weekend.
			child.dateEnd = child.dateStart.AddDays(saveDayCount).SlideOffWeekend()
		}
		child.MaybeShiftDependentsLater()
	}
}

// You want _how many_ days off? No slack time!
const maxAcceptableGapInDays = 3

// MaybeShiftEarlier wants a graph in which an epic starts as soon as possible,
// i.e. right after its tardiest dependency (parent) ends.
func (n *Node) MaybeShiftEarlier() {
	n.cycleCheck()
	if n.seemsDone() {
		return
	}
	minGapDays := 10000 // Assume big
	var tardiest *Node
	for _, parent := range n.dependsOn {
		if gap := parent.dateEnd.DayCount(n.dateStart); gap < minGapDays {
			tardiest = parent
			minGapDays = gap
		}
	}
	if minGapDays < 0 {
		// The child (this) starts before one of its parents ends.
		// Shouldn't be possible after a run of MaybeShiftDependentsLater
	}
	if minGapDays > maxAcceptableGapInDays && tardiest != nil {
		// We can move "this" left in the calendar.
		saveDayCount := n.dateStart.DayCount(n.dateEnd)
		n.dateStart = tardiest.dateEnd.AddDays(maxAcceptableGapInDays).SlideOverWeekend()
		n.dateEnd = n.dateStart.AddDays(saveDayCount).SlideOffWeekend()
	}
	for _, parent := range n.dependsOn {
		parent.MaybeShiftEarlier()
	}
}

// arbitrary
const maxVisitsPerNode = 100

// cycleCheck panics if it suspects a cycle.
// Cycle determination is a bit tricky, because we expect to
// visit children more than once, because a child might have
// multiple 'parents', and we want dates to cascade later in time
// across the graph based on the slowest path.
func (n *Node) cycleCheck() {
	n.visitCount++
	if n.visitCount > maxVisitsPerNode {
		err := fmt.Errorf(
			"visited %q %d times; use 'dot' command to examine graph for cycle, and 'block --remove' to fix",
			n.issue.MyKey, n.visitCount)
		log.Fatal(err)
	}
}
