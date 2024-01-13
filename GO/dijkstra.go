package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)


type Graph struct {
	Nodes []*Node
}

type Node struct {
	Name         string
	Edges        []*Edge
	Channel      chan Message
	RoutingTable map[string]map[string]*Node
}

type Edge struct {
	To     *Node
	Weight int
}

type Message struct {
	Source      *Node
	Destination *Node
	Content     string
}

//definition constantes du graphe
const (
	numNodes = 20 			//taille du graphe 
	minEdgesPerNode = 2		
	maxEdgesPerNode = 5
	weightRange     = 20	//poids max des connections
)

//****		FONCTION CRÉATION GRAPHE ALÉATOIRE ****//
func creation_graph() Graph {

	nodeA := &Node{Name: "A"}
	nodeB := &Node{Name: "B"}
	nodeC := &Node{Name: "C"}
	nodeD := &Node{Name: "D"}
	nodeE := &Node{Name: "E"}
	nodeF := &Node{Name: "F"}
	nodeG := &Node{Name: "G"}
	nodeA.Edges = []*Edge{{To: nodeB, Weight: 1}, {To: nodeC, Weight: 4}, {To: nodeD, Weight: 8}}
	nodeB.Edges = []*Edge{{To: nodeA, Weight: 1}, {To: nodeD, Weight: 2}, {To: nodeG, Weight: 2}}
	nodeC.Edges = []*Edge{{To: nodeA, Weight: 4}, {To: nodeF, Weight: 6}}
	nodeD.Edges = []*Edge{{To: nodeA, Weight: 8}, {To: nodeB, Weight: 2}, {To: nodeE, Weight: 10}, {To: nodeG, Weight: 5}}
	nodeE.Edges = []*Edge{{To: nodeD, Weight: 10}, {To: nodeF, Weight: 7}}
	nodeF.Edges = []*Edge{{To: nodeC, Weight: 6}, {To: nodeE, Weight: 7}, {To: nodeG, Weight: 3}}
	nodeG.Edges = []*Edge{{To: nodeB, Weight: 2}, {To: nodeD, Weight: 5}, {To: nodeF, Weight: 3}}
	graph := Graph{Nodes: []*Node{nodeA, nodeB, nodeC, nodeD, nodeE, nodeF, nodeG}}

	return graph
	// nodeA := &Node{Name: "A"}
	// nodeB := &Node{Name: "B"}
	// nodeC := &Node{Name: "C"}
	// nodeA.Edges = []*Edge{{To: nodeB, Weight: 1}, {To: nodeC, Weight: 4}}
	// nodeB.Edges = []*Edge{{To: nodeA, Weight: 1}, {To: nodeC, Weight: 2}}
	// nodeC.Edges = []*Edge{{To: nodeA, Weight: 4}, {To: nodeB, Weight: 2}}
	// graph := Graph{Nodes: []*Node{nodeA, nodeB, nodeC}}

}

//****	FONCTIONS TRANSMITION DE MESSAGES	****//

func sendMessage(messageChan chan Message, messageEnvoye Message) {
	messageChan <- messageEnvoye
}

func hello(nodeSrc *Node, nodeDst *Node) {
	channel := nodeSrc.Channel
	helloMessage := Message{Source: nodeSrc, Destination: nodeDst, Content: "Hello"}
	sendMessage(channel, helloMessage)
}

func processMessages(messageChan <-chan Message) {
	for {
		select {
		case message := <-messageChan:
			fmt.Printf("Node %s sent message '%s' to Node %s\n", message.Source.Name, message.Content, message.Destination.Name)
		}
	}
}

func routing(node *Node) {
	select {
	case receivedMessage := <-node.Channel:
		if receivedMessage.Destination == node {
			helloAckMessage := Message{Source: receivedMessage.Destination, Destination: receivedMessage.Source, Content: "Hello Ack"}
			next_hop := node.RoutingTable[helloAckMessage.Destination.Name]["next_hop"]
			channel := next_hop.Channel
			sendMessage(channel, helloAckMessage)
		} else {

		}
	}

}


//**** 		FONCTIONS CONSTRUCTION TABLES DE ROUTAGE		****//
func Dijkstra(g *Graph, start *Node) (map[*Node]int, map[*Node]*Node) {
	unvisited := make(map[*Node]struct{})
	distances := make(map[*Node]int)
	next_hop := make(map[*Node]*Node)
	for _, node := range g.Nodes {
		if node == start {
			distances[node] = 0
			next_hop[node] = node
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

//**** 		FONCTION MAIN		 ****//

func main() {

	graph := creation_graph()
	var wg sync.WaitGroup
	// results := make(map[string]map[string]map[string]string, len(graph.Nodes))
	results := make(map[string]map[string]map[string]*Node)
	// print(results)
	for _, start := range graph.Nodes {
		wg.Add(1)
		go func(start *Node) {
			defer wg.Done()
			distances, next_hop := Dijkstra(&graph, start)
			// fmt.Print("next_hop :", next_hop, "\n")
			results[start.Name] = make(map[string]map[string]*Node)
			for node, _ := range distances {
				results[start.Name][node.Name] = make(map[string]*Node)
				results[start.Name][node.Name]["next_hop"] = (next_hop[node]) //.Name
				// results[start.Name][node.Name]["distance"] = fmt.Sprint(dist)
			}
		}(start)
	}
	wg.Wait()

	// Affichage table de routage pour chaque noeud
	for _, start := range graph.Nodes {
		fmt.Println("\nDistances les plus courtes du noeud", start.Name)
		for dest, route := range results[start.Name] {
			fmt.Print(start.Name, " -> ", dest, " : ", route["next_hop"], " -- ", route["distance"], "\n")
		}
	}
}