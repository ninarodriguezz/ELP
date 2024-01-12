package main

import (
	"fmt"
	"sync"
)

type Graph struct {
	Nodes []*Node
}

type Node struct {
	Name  string
	Edges []*Edge
}

type Edge struct {
	To     *Node
	Weight int
}

func Dijkstra(g *Graph, start *Node) (map[*Node]int, map[*Node]*Node) {
	unvisited := make(map[*Node]struct{})
	distances := make(map[*Node]int)
	next_hop := make(map[*Node]*Node)
	for _, node := range g.Nodes {
		if node == start {
			distances[node] = 0
<<<<<<< HEAD
			next_hop[node]= node
=======
			next_hop[node] = node
>>>>>>> 3c2ac86ca3fa2a2bc76ed54ea429737bbb7db717
		} else {
			distances[node] = 1<<31 - 1
		}
		unvisited[node] = struct{}{}
	}

	for len(unvisited) != 0 {
		u := minDist(unvisited, distances)
		if u == nil {
			break
		}
		delete(unvisited, u)
		for _, e := range u.Edges {
			v := e.To
			alt := distances[u] + e.Weight
			if alt < distances[v] {
				distances[v] = alt
				if u == start {
					next_hop[v] = v
				} else {
					next_hop[v] = next_hop[u]
				}
			}
		}
	}
	return distances, next_hop
}

func minDist(unvisited map[*Node]struct{}, distances map[*Node]int) *Node {
	min := 1<<31 - 1
	var n *Node
	for node := range unvisited {
		if distances[node] < min {
			min = distances[node]
			n = node
		}
	}
	return n
}

func main() {
	nodeA := &Node{Name: "A"}
	nodeB := &Node{Name: "B"}
	nodeC := &Node{Name: "C"}
	nodeA.Edges = []*Edge{{To: nodeB, Weight: 1}, {To: nodeC, Weight: 4}}
	nodeB.Edges = []*Edge{{To: nodeA, Weight: 1}, {To: nodeC, Weight: 2}}
	nodeC.Edges = []*Edge{{To: nodeA, Weight: 4}, {To: nodeB, Weight: 2}}
	graph := Graph{Nodes: []*Node{nodeA, nodeB, nodeC}}

	/*    var wg sync.WaitGroup
	    for _, start := range graph.Nodes {
	        wg.Add(1)
	        go func(start *Node) {
	            defer wg.Done()
	            Dijkstra(&graph, start)
	            fmt.Println("Distancias más cortas desde el nodo", start.Name)
	            for _, node := range graph.Nodes {
	                fmt.Printf("%s -> %s: %d\n", start.Name, node.Name, node.Dist)
	            }
	        }(start)
	    }
	    wg.Wait()
	}
	*/

	distances, next_hop := Dijkstra(&graph, nodeA)
	fmt.Print(distances, next_hop)
	fmt.Print("*****")

	var wg sync.WaitGroup
	// results := make(map[string]map[string]map[string]string, len(graph.Nodes))
	results := make(map[string]map[string]map[string]string)
	// print(results)
	for _, start := range graph.Nodes {
		wg.Add(1)
		go func(start *Node) {
			defer wg.Done()
			distances, next_hop := Dijkstra(&graph, start)
			// fmt.Print("next_hop :", next_hop, "\n")
			results[start.Name] = make(map[string]map[string]string)
			for node, dist := range distances {
				results[start.Name][node.Name] = make(map[string]string)
				results[start.Name][node.Name]["next_hop"] = (next_hop[node]).Name
				results[start.Name][node.Name]["distance"] = fmt.Sprint(dist)
			}
		}(start)
	}
	wg.Wait()

	for _, start := range graph.Nodes {
		fmt.Println("\nDistances les plus courtes du noeud", start.Name)
		for dest, route := range results[start.Name] {
			fmt.Print(start.Name, " -> ", dest, " : ", route["next_hop"], " -- ", route["distance"])
		}
	}
}
