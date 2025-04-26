package myj

import (
	"fmt"
	"io"
	"log"

	"github.com/monopole/gojira/internal/utils"
)

const FlagDoIt = "go"

type Graph struct {
	nodes map[MyKey]*Node
	edges map[Edge]bool
}

func (g *Graph) Nodes() map[MyKey]*Node {
	return g.nodes
}

func (g *Graph) Edges() map[Edge]bool {
	return g.edges
}

type Node struct {
	key            MyKey
	status         IssueStatus
	title          string
	originalStart  utils.Date
	originalEnd    utils.Date
	startD         utils.Date
	endD           utils.Date
	assignee       string
	visitCount     int
	dependsOn      []*Node
	isDependedOnBy []*Node
}

func MakeNode(epic *ResponseIssue) *Node {
	return &Node{
		key:      epic.MyKey,
		status:   epic.Status(),
		title:    epic.MySummary(),
		startD:   epic.DateStart(),
		endD:     epic.DateEnd(),
		assignee: epic.AssigneeLdap(),
	}
}

func (n *Node) seemsDone() bool {
	return n.status == IssueStatusClosed ||
		n.status == IssueStatusDone ||
		n.status == IssueStatusClosedWoAction
}

func (n *Node) writeDiGraphNode(w io.Writer) {
	_, _ = fmt.Fprintf(
		w,
		"  %q [label=\"%s\" style=filled fillcolor=%s];\n",
		n.key, n.digraphLabel(), StatusColor(n.status, ColorKindDot))
}

func (n *Node) digraphLabel() string {
	return fmt.Sprintf(
		"%s\n%s %s\n%s",
		n.key,
		func() string {
			dr, err := utils.MakeDayRangeGentle(n.startD, n.endD)
			if err != nil {
				return "(" + utils.Ellipsis(err.Error(), 20) + ")"
			}
			return dr.PrettyRange()
		}(),
		n.assignee,
		utils.ShortLines(n.title))
}

type Edge struct {
	dependent  MyKey
	dependency MyKey
}

func MakeEdge(child, parent MyKey) Edge {
	return Edge{
		dependent:  child,
		dependency: parent,
	}
}

func MakeGraph(nodes map[MyKey]*Node, edges map[Edge]bool) *Graph {
	g := &Graph{nodes: nodes, edges: edges}
	g.loadEdgesIntoNodes()
	return g
}

// loadEdgesIntoNodes just makes it easier to compute in-degree/out-degree and
// traverse edges, rather than repeatedly scanning all edges looking for node
// matches.
func (g *Graph) loadEdgesIntoNodes() {
	for edge := range g.edges {
		g.nodes[edge.dependent].dependsOn = append(
			g.nodes[edge.dependent].dependsOn, g.nodes[edge.dependency])
		g.nodes[edge.dependency].isDependedOnBy = append(
			g.nodes[edge.dependency].isDependedOnBy, g.nodes[edge.dependent])
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
	for _, node := range g.Nodes() {
		node.writeDiGraphNode(w)
	}
	for edge := range g.Edges() {
		_, _ = fmt.Fprintf(w, "%q -> %q;\n", edge.dependency, edge.dependent)
	}
	_, _ = fmt.Fprintln(w, "}")
}

func (g *Graph) ReportMisOrdering(w io.Writer) {
	for edge := range g.Edges() {
		start := g.nodes[edge.dependent].startD
		end := g.nodes[edge.dependency].endD
		if !start.After(end) {
			_, _ = fmt.Fprintf(w,
				"%10s depends on %10s, but%4d starts on %s, %3d days before%4d ends on %s.\n",
				edge.dependent, edge.dependency,
				edge.dependent.Num, start.Brief(), start.DayCount(end),
				edge.dependency.Num, end.Brief())
		}
	}
}

func (g *Graph) ReportWeekends(w io.Writer) {
	for _, n := range g.Nodes() {
		if n.startD.IsWeekend() {
			_, _ = fmt.Fprintf(w, "%12s starts on a %s (%s), pushing to Mon.\n",
				n.key, n.startD.Weekday(), n.startD.Brief())
			n.startD = n.startD.SlideOverWeekend()
		}
		if n.endD.IsWeekend() {
			_, _ = fmt.Fprintf(w, "%12s ends on a %s (%s), pulling to Fri.\n",
				n.key, n.endD.Weekday(), n.endD.Brief())
			n.endD = n.endD.SlideBeforeWeekend()
		}
	}
}

func (g *Graph) ScanAndReportNodes(w io.Writer) {
	for key := range g.nodes {
		node := g.nodes[key]
		dr, dateErr := utils.MakeDayRangeGentle(node.startD, node.endD)
		if dateErr != nil {
			node.startD = dr.Start()
			node.endD = dr.End()
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
	for _, node := range g.Nodes() {
		node.visitCount = 0
	}
}

func (g *Graph) SaveOriginalDates() {
	for _, node := range g.Nodes() {
		node.originalStart = node.startD
		node.originalEnd = node.endD
	}
}

// MaybeShiftDependentsLater might push dependent epics out in time to
// start after their dependencies end.
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

// arbitrary
const maxVisitsPerNode = 200

// MaybeShiftDependentsLater wants a graph in which no child starts
// before the parent ends.
// I.e. it shifts dependents later if they start before
// their dependency completes.
func (n *Node) MaybeShiftDependentsLater() {
	n.cycleCheck()
	for _, child := range n.isDependedOnBy {
		if child.seemsDone() {
			continue
		}
		if !child.startD.After(n.endD) {
			// This is a problem.

			// A given child may have more than one parent in a digraph,
			// e.g. where several epics block, say, the final epic.
			// Each of these 'parent' epics might have a different completion
			// date. We visit the child once for each parent, moving
			// its start date further out to begin after whichever parent
			// ends the latest.

			saveDayCount := child.startD.DayCount(child.endD)

			// Move child start to one day after parent end.
			child.startD = n.endD.AddDays(1).SlideOverWeekend()

			// Move child end to roughly establish the same duration
			// as before, modulo not ending on a weekend.
			child.endD = child.startD.AddDays(saveDayCount).SlideOffWeekend()
		}
		child.MaybeShiftDependentsLater()
	}
}

// You want _how many_ days off? No slack time!
const maxAcceptableGapInDays = 3

// MaybeShiftEarlier wants a graph in which an epic starts as soon as possible,
// i.e. right after its tardiest dependency ends.
func (n *Node) MaybeShiftEarlier() {
	n.cycleCheck()
	if n.seemsDone() {
		return
	}
	minGapDays := 10000 // Assume big
	var tardiest *Node
	for _, parent := range n.dependsOn {
		if gap := parent.endD.DayCount(n.startD); gap < minGapDays {
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
		saveDayCount := n.startD.DayCount(n.endD)
		n.startD = tardiest.endD.AddDays(maxAcceptableGapInDays).SlideOverWeekend()
		n.endD = n.startD.AddDays(saveDayCount).SlideOffWeekend()
	}
	for _, parent := range n.dependsOn {
		parent.MaybeShiftEarlier()
	}
}

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
			n.key.String(), n.visitCount)
		log.Fatal(err)
	}
}
