package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var k = 200

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
		wg:    sync.WaitGroup{},
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

func Producer(source chan<- Package, g *Graph, k int) {
	g.vertices[0].wg.Add(1)
	rand.Seed(time.Now().UnixNano())
	wg := &g.vertices[0].wg
	vis := make([]int, 0)
	for i := 1; i <= k; i++ {
		m := Package{i, vis}
		fmt.Println("Puting package ", i, " into source...")
		source <- m
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(50)))
	}
	defer wg.Done()
}

func Consumer(link <-chan Package, done chan<- bool) {
	rand.Seed(time.Now().UnixNano())
	for k != 0 {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
		p := <-link
		fmt.Println("Package nr ", p.id, " recieved: \n", p)
		k--
	}
	done <- true
}

func main() {
	var n, d int
	fmt.Println("Type number of vertices: ")
	fmt.Scanf("%d", &n)
	fmt.Println("Type number extra edges: ")
	fmt.Scanf("%d", &d)
	fmt.Println("Type number packages: ")
	fmt.Scanf("%d", &k)

	graph := New()
	//const n = 100
	//const d = 200
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
	//Put extra edges
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < d; i++ {
		v1 := rand.Intn(n - 2)
		v2 := rand.Intn(n-1-v1) + v1 + 1
		graph.AddEdge(v1, v2)
	}

	source := make(chan Package)
	end := make(chan Package)
	endWG := sync.WaitGroup{}
	done := make(chan bool)

	graph.vertices[0].in = source
	//Exit from ending vertex
	graph.vertices[len(graph.vertices)-1].outs = append(graph.vertices[len(graph.vertices)-1].outs, &end)
	graph.vertices[len(graph.vertices)-1].wgs = append(graph.vertices[len(graph.vertices)-1].wgs, &endWG)

	go Producer(source, graph, k)

	for i := 0; i < n; i++ {
		fmt.Println("Starting vertex nr: ", i)
		graph.vertices[i].wg.Add(1)
		go Forwarder(graph.vertices[i])

	}
	go Consumer(end, done)
	//for i := 0; i < n; i++ {
	//	graph.vertices[i].wg.Wait()
	//}
	<-done
}

func Forwarder(vertex *Vertex) {

	in := vertex.in

	for k != 0 {
		//Recieve
		p := <-in
		fmt.Println("  Vertex ", vertex.index, "\n\tPackage recieved: ", p.id)
		p.visited = append(p.visited, vertex.index)
		vertex.packages = append(vertex.packages, p.id)

		//Choose
		rand.Seed(time.Now().UnixNano())
		forwardTo := rand.Intn(len(vertex.outs))
		out := *vertex.outs[forwardTo]
		//out := *vertex.outs[0]
		//Sleep
		ms := rand.Intn(50)
		time.Sleep(time.Duration(ms) * time.Millisecond)

		//Send
		vertex.wgs[forwardTo].Add(1)
		out <- p
		vertex.wgs[forwardTo].Done()
	}
}
