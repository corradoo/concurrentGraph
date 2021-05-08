package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var k = 10

type Graph struct {
	vertices []*Vertex
}

type Vertex struct {
	index    int
	in       chan Package
	packages []int
	outs     []*chan Package
	print    *chan string
	edges    []int
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

func (g *Graph) AddNode(p *chan string) (id int) {
	id = len(g.vertices)
	g.vertices = append(g.vertices, &Vertex{
		index: id,
		in:    make(chan Package),
		print: p,
	})
	return
}

func (g *Graph) AddEdge(v1, v2 int) {
	//Adding channel to out slice
	g.vertices[v1].outs = append(g.vertices[v1].outs, &g.vertices[v2].in)
	g.vertices[v1].edges = append(g.vertices[v1].edges, g.vertices[v2].index)
}

func (g *Graph) Nodes() []int {
	vertices := make([]int, len(g.vertices))
	for i := range g.vertices {
		vertices[i] = i
	}
	return vertices
}

func Producer(source chan<- Package, k int, printer *chan string) {
	rand.Seed(time.Now().UnixNano())
	vis := make([]int, 0)
	for i := 1; i <= k; i++ {
		m := Package{i, vis}
		msg := fmt.Sprint("Putting package ", i, " into source...")
		*printer <- msg
		source <- m
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(50)))
	}
}

func Consumer(link <-chan Package, wg *sync.WaitGroup, packages *[]*Package, printer *chan string) {
	rand.Seed(time.Now().UnixNano())
	for k != 0 {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
		p := <-link
		*packages = append(*packages, &p)
		msg := fmt.Sprint("Package nr ", p.id, " received:\n\t", p)
		*printer <- msg
		//fmt.Println("Package nr ", p.id, " received: \n", p)
		k--
	}
	wg.Done()
}

func Forwarder(vertex *Vertex) {

	in := vertex.in

	for {
		//Receive
		p := <-in
		//fmt.Println("  Vertex ", vertex.index, "\n\tPackage received: ", p.id)
		msg := "  Vertex " + strconv.Itoa(vertex.index) + "\n\tPackage received: " + strconv.Itoa(p.id)

		*vertex.print <- msg
		p.visited = append(p.visited, vertex.index)
		vertex.packages = append(vertex.packages, p.id)

		//Choose
		rand.Seed(time.Now().UnixNano())
		forwardTo := rand.Intn(len(vertex.outs))
		out := *vertex.outs[forwardTo]

		//Sleep
		ms := rand.Intn(50)
		time.Sleep(time.Duration(ms) * time.Millisecond)

		//Send
		out <- p
	}
}

func Printer(print <-chan string, pWg *sync.WaitGroup) {
	for msg := range print {
		fmt.Println(msg)
	}
	fmt.Print("LOG END ---------------------------------------------------------------")
	pWg.Done()
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

	nodes := make([]int, n)
	printChan := make(chan string, 100)
	packages := make([]*Package, 0)

	//Make nodes
	for i := 0; i < n; i++ {
		nodes[i] = graph.AddNode(&printChan)
	}

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
	wg := sync.WaitGroup{}
	pWg := sync.WaitGroup{}

	go Printer(printChan, &pWg)
	for i, v := range graph.vertices {
		msg := fmt.Sprint("\t\tVertex ", i, v.edges)
		printChan <- msg
	}

	graph.vertices[0].in = source

	//Exit from ending vertex to Consumer
	graph.vertices[len(graph.vertices)-1].outs = append(graph.vertices[len(graph.vertices)-1].outs, &end)
	go Producer(source, k, &printChan)

	for i := 0; i < n; i++ {
		go Forwarder(graph.vertices[i])
	}

	wg.Add(1)
	go Consumer(end, &wg, &packages, &printChan)

	//Wait till all packages reach Consumer
	wg.Wait()

	//Final report
	pWg.Add(1)
	printChan <- "Final:"
	printChan <- "\tVertices:"
	for i, v := range graph.vertices {
		msg := fmt.Sprint("\t\tVertex ", i, v.packages)
		printChan <- msg
	}
	printChan <- "\tPackages:"
	for _, v := range packages {
		msg := fmt.Sprint("\t\tPackage ", *v)
		printChan <- msg
	}
	close(printChan)
	pWg.Wait()
}
