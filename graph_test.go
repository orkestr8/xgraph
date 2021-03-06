package xgraph // import "github.com/orkestr8/xgraph"

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

func TestGonumGraph(t *testing.T) {

	g := simple.NewDirectedGraph()

	// AddNode must be called right after NewNode to ensure the ID is properly assigned and registered
	// in the graph, or we'd get ID collision panic.
	a := g.NewNode()
	require.Nil(t, g.Node(a.ID()))
	g.AddNode(a)
	require.NotNil(t, g.Node(a.ID()))

	b := g.NewNode()
	g.AddNode(b)

	aLikesB := g.NewEdge(a, b)
	require.Nil(t, g.Edge(a.ID(), b.ID()))
	g.SetEdge(aLikesB)

	cycle := topo.DirectedCyclesIn(g)
	require.Equal(t, 0, len(cycle))

	// Calling ReversedEdge doesn't actually reverses the edge in the graph.
	reversed := aLikesB.ReversedEdge()
	require.Nil(t, g.Edge(b.ID(), a.ID()))
	require.NotNil(t, g.Edge(a.ID(), b.ID()))

	// Now an edge exists.  For this DAG we have a loop now.
	g.SetEdge(reversed)
	require.NotNil(t, g.Edge(b.ID(), a.ID()))
	require.NotNil(t, g.Edge(a.ID(), b.ID()))

	_, err := topo.SortStabilized(g, nil)
	require.Error(t, err)

	cycle = topo.DirectedCyclesIn(g)
	require.Equal(t, 1, len(cycle))
	t.Log(cycle)

	c := g.NewNode()
	g.AddNode(c)
	g.SetEdge(g.NewEdge(a, c))
	g.SetEdge(g.NewEdge(c, a))
	cycle = topo.DirectedCyclesIn(g)
	require.Equal(t, 2, len(cycle))
	t.Log(cycle)
}

func TestAdd(t *testing.T) {

	A := &nodeT{id: "A"}
	B := &nodeT{id: "B"}
	C := &nodeT{id: "C"}
	plus := &nodeT{id: "+"}
	minus := &nodeT{id: "-"}

	g := Builder(Options{})
	require.NoError(t, g.Add(A, B, C, plus, minus))

	require.Error(t, g.Add(&nodeT{id: "A"}), "Not OK for duplicate key when struct identity fails")
	require.NoError(t, g.Add(A), "Idempotent: same node by identity")

	for _, n := range []Node{plus, minus, A, B, C} {
		require.NotNil(t, g.Node(n.NodeKey()))
	}

	// We should expect the references to be the same
	require.Equal(t, A, g.Node("A"))
	require.Equal(t, B, g.Node("B"))
	require.Equal(t, C, g.Node("C"))
}

func TestAssociate(t *testing.T) {

	A := &nodeT{id: "A"}
	B := &nodeT{id: "B"}
	C := &nodeT{id: "C"}
	D := &nodeT{id: "D"}

	g := Builder(Options{})
	require.NoError(t, g.Add(A, B, C))

	require.NotNil(t, g.Node(A.NodeKey()))
	require.NotNil(t, g.Node(B.NodeKey()))
	require.NotNil(t, g.Node(C.NodeKey()))
	require.Nil(t, g.Node(D.NodeKey()))

	likes := EdgeKind(1)
	shares := EdgeKind(2)

	ev, err := g.Associate(A, likes, B, Attribute{Key: "key1", Value: "some context"},
		Attribute{Key: "key2", Value: "something else"})
	require.NoError(t, err)
	require.NotNil(t, g.Edge(A, likes, B))
	require.NotNil(t, ev)
	require.Equal(t, A, ev.From())
	require.Equal(t, B, ev.To())
	require.Equal(t, map[string]interface{}{
		"key1": "some context",
		"key2": "something else"}, ev.Attributes())

	require.Equal(t, A, g.Edge(A, likes, B).From())
	require.Equal(t, B, g.Edge(A, likes, B).To())
	require.Equal(t, map[string]interface{}{
		"key1": "some context",
		"key2": "something else"}, g.Edge(A, likes, B).Attributes())

	_, err = g.Associate(D, likes, A)
	require.Error(t, err, "Expects error because D was not added to the graph.")
	require.Nil(t, g.Edge(D, likes, A), "Expects false because C is not part of the graph.")

	_, err = g.Associate(A, likes, C)
	require.NoError(t, err, "No error because A and C are members of the graph.")
	require.NotNil(t, g.Edge(A, likes, C), "A likes C.")
	require.Equal(t, 0, len(g.Edge(A, likes, C).Attributes()))
	require.Nil(t, g.Edge(C, shares, A), "Shares is not an association kind between A and B.")

	// Repeated calls to get the edge always result in the same reference:
	edge1 := g.Edge(A, likes, C)
	edge2 := g.Edge(A, likes, C)
	require.True(t, edge1 == edge2)
	lookup := map[Edge]interface{}{
		edge1: 1,
	}
	require.Equal(t, 1, lookup[edge1])
	require.Equal(t, 1, lookup[edge2])

	require.Equal(t, 2, len(g.From(A, likes).Edges().Slice()))
	require.Equal(t, 1, len(g.To(likes, B).Edges().Slice()))
	require.Equal(t, "A", g.To(likes, B).Edges().Slice()[0].From().NodeKey())
	require.Equal(t, "B", g.To(likes, B).Edges().Slice()[0].To().NodeKey())
	require.Equal(t, 1, len(g.To(likes, C).Edges().Slice()))
	require.Equal(t, "A", g.To(likes, C).Edges().Slice()[0].From().NodeKey())
	require.Equal(t, "C", g.To(likes, C).Edges().Slice()[0].To().NodeKey())
	require.Equal(t, 0, len(g.From(B, likes).Edges().Slice()))
	require.Equal(t, 0, len(g.From(C, likes).Edges().Slice()))
	require.Equal(t, 0, len(g.To(likes, A).Edges().Slice()), "D was not added")
	require.Equal(t, 0, len(g.From(D, likes).Edges().Slice()), "D was not added")

}
