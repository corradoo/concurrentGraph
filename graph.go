package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var k = 10
var h = 30

type Edge struct {
	v1 int
	v2 int
}

type Graph struct {
	vertices []*Vertex
}

type Vertex struct {
	index      int
	in         chan Package
	packages   []int
	vertices   []*Vertex
	print      *chan string
	hasPacket  bool
	gateChan   chan bool
	hunterChan chan bool
	isTrapped  bool
	thrash     *chan Package
	rounds     int
}

type Package struct {
	id      int
	visited []int
	ttl     int
}

func New() *Graph {
	return &Graph{
		vertices: []*Vertex{},
	}
}

func (g *Graph) AddNode(p *chan string, t *chan Package) (id int) {
	id = len(g.vertices)
	g.vertices = append(g.vertices, &Vertex{
		index:      id,
		in:         make(chan Package),
		print:      p,
		gateChan:   make(chan bool),
		thrash:     t,
		hunterChan: make(chan bool),
		rounds:     0,
	})
	return
}

func (g *Graph) AddEdge(v1, v2 int) {
	//Adding channel to out slice
	g.vertices[v1].vertices = append(g.vertices[v1].vertices, g.vertices[v2])
	/*g.vertices[v1].outs = append(g.vertices[v1].outs, &g.vertices[v2].in)
	g.vertices[v1].edges = append(g.vertices[v1].edges, g.vertices[v2].index)*/
}

func (g *Graph) Nodes() []int {
	vertices := make([]int, len(g.vertices))
	for i := range g.vertices {
		vertices[i] = i
	}
	return vertices
}

func Producer(vertex *Vertex, k int, printer *chan string) {
	rand.Seed(time.Now().UnixNano())
	vis := make([]int, 0)
	for i := 1; i <= k; i++ {
		p := Package{i, vis, h}
		for {
			vertex.gateChan <- true
			res := <-vertex.gateChan
			*printer <- fmt.Sprint("source response: ", res)
			if res {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(50)))
			} else {
				msg := fmt.Sprint("Putting package ", i, " into source...")
				*printer <- msg
				vertex.in <- p
				break
			}
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(50)))
		}
	}
}

func Consumer(vertex *Vertex, wg *sync.WaitGroup, packages *[]*Package, printer *chan string) {
	rand.Seed(time.Now().UnixNano())
	for k != 0 {
		//time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
		p := <-vertex.in
		p.visited = append(p.visited, vertex.index)
		*packages = append(*packages, &p)
		msg := fmt.Sprint("Package nr ", p.id, " received:\n\t", p)
		*printer <- msg
		k--
		vertex.hasPacket = false
	}
	wg.Done()
}

func Forwarder(vertex *Vertex) {

	in := vertex.in
	hun := vertex.hunterChan
	forwardTo := vertex.vertices[0].in
	for {
		select {
		//Receive
		case p := <-in:
			msg := "  Vertex " + strconv.Itoa(vertex.index) + "\n\tPackage received: " + strconv.Itoa(p.id)
			*vertex.print <- msg
			p.visited = append(p.visited, vertex.index)
			vertex.packages = append(vertex.packages, p.id)

			//Check whether vertex is trapped
			if vertex.isTrapped {
				msg := "  Vertex " + strconv.Itoa(vertex.index) + "\n\tPackage CAUGHT by HUNTER: " + strconv.Itoa(p.id)
				*vertex.print <- msg
				*vertex.thrash <- p
				vertex.isTrapped = false
			} else if p.ttl == 0 { // Check if package can live
				msg := "  Vertex " + strconv.Itoa(vertex.index) + "\n\tPackage DIED: " + strconv.Itoa(p.id)
				*vertex.print <- msg
				*vertex.thrash <- p
			} else { //Package can be forwarded
				//*vertex.print <- fmt.Sprint("Chcę wysłać")
				p.ttl--
				//*vertex.print <- fmt.Sprint("TTL: ", p.ttl)
				//Sleep
				ms := rand.Intn(50)
				time.Sleep(time.Duration(ms) * time.Millisecond)

				//Try to send
				rand.Seed(time.Now().UnixNano())

			Round:
				for {
					*vertex.print <- fmt.Sprint("\t\tVertex ", vertex.index, " shuffling...")
					rand.Shuffle(len(vertex.vertices), func(i, j int) {
						vertex.vertices[i], vertex.vertices[j] = vertex.vertices[j], vertex.vertices[i]
					})
					vertex.rounds++
					for i, out := range vertex.vertices {
						out.gateChan <- true
						res := <-out.gateChan
						*vertex.print <- fmt.Sprint("\t\tVertex ", vertex.index, " Response from keeper(V:", vertex.vertices[i].index, ") ", res)
						if res { //Blocked, lipa
							continue
						} else {
							forwardTo = out.in
							*vertex.print <- fmt.Sprint("\t\tVertex ", vertex.index, " Chosen: ", out.index)
							break Round
						}
					}
					time.Sleep(time.Duration(ms) * time.Millisecond)
				}
				forwardTo <- p
			}
			vertex.hasPacket = false

		//Place trap
		case <-hun:
			vertex.isTrapped = true

		default:
			continue
		}

	}
}

func Gatekeeper(vertex *Vertex) {

	for {
		<-vertex.gateChan
		if vertex.hasPacket {
			vertex.gateChan <- true
		} else {
			vertex.hasPacket = true
			vertex.gateChan <- false
		}
		time.Sleep(time.Millisecond * 5)
	}
}

func Printer(print <-chan string, pWg *sync.WaitGroup) {
	for msg := range print {
		fmt.Println(msg)
	}
	fmt.Print("LOG END ---------------------------------------------------------------")
	pWg.Done()
}

func Trasher(t <-chan Package, wg *sync.WaitGroup, packages *[]*Package) {
	for k != 0 {
		p := <-t
		fmt.Println(p, " In TRASH")
		*packages = append(*packages, &p)
		k--
	}
	wg.Done()
}

func Hunter(g *Graph) {

	rand.Seed(time.Now().UnixNano())

	for {
		time.Sleep(time.Millisecond * 500)
		v := g.vertices[rand.Intn(len(g.vertices))]
		v.hunterChan <- true
	}
}

func main() {
	var n, d, b = 10, 35, 35
	/*fmt.Println("Type number of vertices: ")
	fmt.Scanf("%d", &n)
	fmt.Println("Type number extra forward edges: ")
	fmt.Scanf("%d", &d)
	fmt.Println("Type number extra backward edges: ")
	fmt.Scanf("%d", &b)
	fmt.Println("Type number packages: ")
	fmt.Scanf("%d", &k)
	fmt.Println("Type TTL for all packs: ")
	fmt.Scanf("%d", &h)*/

	binomial := new(big.Int)
	binomial.Binomial(int64(n), 2)
	edges := int64(d + b + n - 1)
	fmt.Println("Edges percentage: ", float64(edges)/(float64(binomial.Int64())*2.0))
	graph := New()

	nodes := make([]int, n)
	printChan := make(chan string, 100)
	trashChan := make(chan Package)
	packages := make([]*Package, 0)

	//Make nodes
	for i := 0; i < n; i++ {
		nodes[i] = graph.AddNode(&printChan, &trashChan)
	}

	//Put 'normal' edges
	for i := 0; i < n-1; i++ {
		graph.AddEdge(nodes[i], nodes[i+1])
	}

	//Put extra edges
	rand.Seed(time.Now().UnixNano())

	//Make all possible edges, and then shuffle them
	forwardEdges := make([]Edge, 0)
	backwardEdges := make([]Edge, 0)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if i != j {
				if i+1 < j {
					forwardEdges = append(forwardEdges,
						Edge{
							v1: i,
							v2: j,
						})
				}
				backwardEdges = append(backwardEdges,
					Edge{
						v1: j,
						v2: i,
					})
			}
		}
	}

	rand.Shuffle(len(forwardEdges), func(i, j int) {
		forwardEdges[i], forwardEdges[j] = forwardEdges[j], forwardEdges[i]
	})
	rand.Shuffle(len(backwardEdges), func(i, j int) {
		backwardEdges[i], backwardEdges[j] = backwardEdges[j], backwardEdges[i]
	})

	fmt.Println(backwardEdges)
	fmt.Println(forwardEdges)

	for i := 0; i < d; i++ {
		v1 := forwardEdges[i].v1
		v2 := forwardEdges[i].v2
		graph.AddEdge(v1, v2)
	}

	for i := 0; i < b; i++ {
		v1 := backwardEdges[i].v1
		v2 := backwardEdges[i].v2
		graph.AddEdge(v1, v2)
	}

	//source := make(chan Package)
	end := &Vertex{
		index:     -1,
		in:        make(chan Package),
		print:     &printChan,
		hasPacket: false,
		gateChan:  make(chan bool),
	}
	wg := sync.WaitGroup{}
	pWg := sync.WaitGroup{}
	graph.vertices[len(graph.vertices)-1].vertices = append(graph.vertices[len(graph.vertices)-1].vertices, end)

	go Printer(printChan, &pWg)

	for i, v := range graph.vertices {
		msg := fmt.Sprint("\t\tVertex ", i, " [ ")
		for _, outs := range v.vertices {
			msg = msg + fmt.Sprint(outs.index, " ")
		}
		msg = msg + "]"
		printChan <- msg
	}
	go Trasher(trashChan, &wg, &packages)
	//graph.vertices[0].in = source

	//Exit from ending vertex to Consumer

	go Producer(graph.vertices[0], k, &printChan)

	for i := 0; i < n; i++ {
		go Gatekeeper(graph.vertices[i])
		go Forwarder(graph.vertices[i])
	}

	wg.Add(1)

	go Gatekeeper(end)
	go Consumer(end, &wg, &packages, &printChan)
	go Hunter(graph)
	//Wait till all packages reach Consumer / get deleted
	wg.Wait()

	//Final report
	pWg.Add(1)
	printChan <- "Final:"
	printChan <- "\tVertices:"
	for i, v := range graph.vertices {
		msg := fmt.Sprint("\t\tVertex ", i, v.packages, " packs: ", len(v.packages), " rounds: ", v.rounds)
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
