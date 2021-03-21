package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Graph struct {
	vertices []*Vertex
}

type Vertex struct {
	index    int
	in       chan Package
	wg       sync.WaitGroup
	packages []int
	outs     []*chan Package
	wgs      []*sync.WaitGroup
}

type Package struct {
	id      int
	visited []int
}

func New() *Graph {
	return &Graph{
		vertices: []*Vertex{},
	}
}

func (g *Graph) AddNode() (id int) {
	id = len(g.vertices)
	g.vertices = append(g.vertices, &Vertex{
		index: id,
		in:    make(chan Package),
	})
	return
}

func (g *Graph) AddEdge(v1, v2 int) {
	//Adding channel to out slice
	g.vertices[v1].outs = append(g.vertices[v1].outs, &g.vertices[v2].in)
	g.vertices[v1].wgs = append(g.vertices[v1].wgs, &g.vertices[v2].wg)
}

func (g *Graph) Nodes() []int {
	vertices := make([]int, len(g.vertices))
	for i := range g.vertices {
		vertices[i] = i
	}
	return vertices
}

//func (g *Graph) Edges() [][2]int {
//	edges := make([][2]int, 0, len(g.vertices))
//	for i := 0; i < len(g.vertices); i++ {
//		for k := range g.vertices[i].edges {
//			edges = append(edges, [2]int{i, k})
//		}
//	}
//	return edges
//}

func Producer(link chan<- Package, g *Graph) {

	wg := &g.vertices[0].wg
	wg.Add(1)
	defer wg.Done()
	vis := make([]int, 0)
	m := Package{0, vis}
	link <- m
}

func Consumer(link <-chan Package, done chan<- bool) {
	fmt.Println(<-link)
	done <- true
}

func main() {
	graph := New()
	const n = 3
	//d:=20
	nodes := make([]int, n)
	//Make nodes
	for i := 0; i < n; i++ {
		nodes[i] = graph.AddNode()
	}
	fmt.Println("Nodes: ", graph.Nodes())

	//Put 'normal' edges
	for i := 0; i < n-1; i++ {
		graph.AddEdge(nodes[i], nodes[i+1])
	}

	//fmt.Println("Edges: ", graph.Edges())

	//Put extra edges
	//rand.Seed(time.Now().UnixNano())
	//for i:=0; i<d; i++ {
	//	v1:= rand.Intn(n-2)
	//	v2:= rand.Intn(n-1-v1)+v1+1
	//	graph.AddEdge(v1,v2)
	//}

	//fmt.Println("Extended edges: ", graph.Edges())
	source := make(chan Package)
	end := make(chan Package)
	done := make(chan bool)

	//Exit from ending vertex
	graph.vertices[len(graph.vertices)-1].outs = append(graph.vertices[len(graph.vertices)-1].outs, &end)

	go Producer(source, graph)
	go Consumer(end, done)
	for i := 0; i < n; i++ {
		graph.vertices[i].wg.Add(1)
		fmt.Println("Starting vertex nr: ", i)
		go Forwarder(graph.vertices[i])

	}

	<-done
}

func Forwarder(vertex *Vertex) {
	defer vertex.wg.Done()
	in := vertex.in
	//Recieve
	p := <-in
	fmt.Println("Package recieved: ", p.id)
	p.visited = append(p.visited, vertex.index)
	vertex.packages = append(vertex.packages, p.id)

	//Choose
	rand.Seed(time.Now().UnixNano())
	//out := vertex.outs[rand.Intn(len(vertex.outs))]
	out := *vertex.outs[0]
	//Sleep
	time.Sleep(time.Duration(rand.Intn(1000)))

	//Send
	out <- p
}