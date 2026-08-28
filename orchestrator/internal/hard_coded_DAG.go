package internal

import "github.com/google/uuid"

type Node struct {
	Name string
	ID   uuid.UUID
}

type Graph struct {
	Nodes map[string]*Node //names should be unique
	Edges map[uuid.UUID][]uuid.UUID
}

func CreateGraph() *Graph {
	edges := make(map[uuid.UUID][]uuid.UUID)
	nodes := make(map[string]*Node)
	g := Graph{
		Nodes: nodes,
		Edges: edges,
	}
	g.createNode("A")
	g.createNode("B")
	g.createNode("C")
	g.createNode("D")
	g.createNode("E")
	g.createNode("F")
	g.createNode("G")

	return &g
}

func (g *Graph) HardCodeIt() {
	g.addEdge("A", "C")
	g.addEdge("B", "C")
	g.addEdge("C", "D")
	g.addEdge("E", "D")
	g.addEdge("G", "A")
	g.addEdge("G", "B")
}

func (g *Graph) createNode(name string) {
	node := Node{
		Name: name,
	}
	g.Nodes[name] = &node
}

func (g *Graph) addEdge(parent string, child string) {
	p_id := g.Nodes[parent]
	c_id := g.Nodes[child]
	g.Edges[p_id.ID] = append(g.Edges[p_id.ID], c_id.ID)

}
