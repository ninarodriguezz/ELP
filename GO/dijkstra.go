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
	minEdgesPerNode = 2		//au moins deux pour s'assurer qu'un node n'est pas isolé, probabilité de configuration de trois noeuds en triangle negligé :P 
	maxEdgesPerNode = 5
	weightRange     = 20	//poids max des connections
)

//****		FONCTION CRÉATION GRAPHE ALÉATOIRE ****//
func initRandomGraph(nodesCount int) Graph {
	rand.Seed(time.Now().UnixNano())			
	//permet d'obtenir une séquence aléatoire differente à chaque execution du code, 
	//on se base sur le temps qui est un parametre que change constantement 

	// On appele les nodes A + num comme ça il y a pas de confusion avec les poids et on a inf possibilités
	nodes := make([]*Node, nodesCount)
	for i := 0; i < nodesCount; i++ {
		nodes[i] = &Node{Name: fmt.Sprintf("A%d", i+1)}
	}

	// Crear conexiones aleatorias entre nodos
	for _, node := range nodes {
		// Determiner aléatoriamente la quantité d'Edges que le node aura (n entre minEdgesPerNode et maxEdgesPerNode)
		edgesCount := rand.Intn(maxEdgesPerNode-minEdgesPerNode+1) + minEdgesPerNode

		// Crear edges
		for j := 0; j < edgesCount; j++ {
			// Choisir un node aléatoire
			otherNode := nodes[rand.Intn(nodesCount)]

			// Evitar conexión consigo mismo y duplicados
			for node == otherNode || edgeExists(node, otherNode) {
				otherNode = nodes[rand.Intn(nodesCount)]
			}

			// Crear bidireccionalidad
			edge := &Edge{To: otherNode, Weight: rand.Intn(weightRange)}
			node.Edges = append(node.Edges, edge)
			otherNode.Edges = append(otherNode.Edges, &Edge{To: node, Weight: edge.Weight})
		}
	}

	return Graph{Nodes: nodes}
}

// Verificar si ya existe un edge entre dos nodos
func edgeExists(nodeA, nodeB *Node) bool {
	for _, edge := range nodeA.Edges {
		if edge.To == nodeB {
			return true
		}
	}
	return false
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

	graph := createRandomGraph(100) //en parametre on met le nombre de sommets
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