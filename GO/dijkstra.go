package main

import (
    "fmt"
    "sync"
)

type Graph struct {
    Nodes []*Node
}

type Node struct {
    Name   string
    Edges  []*Edge
    Visited bool
    Dist   int
}

type Edge struct {
    To     *Node
    Weight int
}

func Dijkstra(g *Graph, start *Node) map[*Node]int {
    unvisited := make(map[*Node]struct{})
    distances := make(map[*Node]int)
    for _, node := range g.Nodes {
        if node == start {
            distances[node] = 0
        } else {
            distances[node] = 1<<31 - 1
        }
        unvisited[node] = struct{}{}
    }

    for len(unvisited) != 0 {
        u := minDist(unvisited)
        if u == nil {
            break
        }
        delete(unvisited, u)
        for _, e := range u.Edges {
            v := e.To
            alt := distances[u] + e.Weight
            if alt < distances[v] {
                distances[v] = alt
            }
        }
    }
    return distances
}

func minDist(unvisited map[*Node]struct{}) *Node {
    min := 1<<31 - 1
    var n *Node
    for node := range unvisited {
        if node.Dist < min {
            min = node.Dist
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

    var wg sync.WaitGroup
    results := make([]map[string]int, len(graph.Nodes))
    print(results)
    for i, start := range graph.Nodes {
        wg.Add(1)
        go func(i int, start *Node) {
            defer wg.Done()
            distances := Dijkstra(&graph, start)
            results[i] = make(map[string]int)
            for node, dist := range distances {
                results[i][node.Name] = dist
            }
        }(i, start)
    }
    wg.Wait()

    for i, start := range graph.Nodes {
        fmt.Println("Distancias más cortas desde el nodo", start.Name)
        for name, dist := range results[i] {
            fmt.Printf("%s -> %s: %d\n", start.Name, name, dist)
        }
    }
}
