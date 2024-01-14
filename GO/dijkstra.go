package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// Structure définissant un graphe
type Graph struct {
	Nodes []*Node
}

// Structure définissant un nœud dans le graphe
type Node struct {
	Name         string
	Edges        []*Edge
	Channel      chan Message
	RoutingTable map[string]map[string]*Node //Table de routage de chaque node qui contient tous les autres sommets avec le next_hop (sans distance)
}

// Structure définissant une arête reliant deux nœuds
type Edge struct {
	To     *Node
	Weight int
}

// Structure définissant un message envoyé entre nœuds
type Message struct {
	Source      *Node
	Destination *Node
	Content     string
	LinkDetails LinkInfo //info contenue dans les messages pour informer qu'on a perdu ou établi un nouveau lien
}

type LinkInfo struct {
	NodeA *Node
	NodeB *Node //nodes qui ont perdu ou récuperé un lien
}

// definition caracteristiques du graphe
const (
	minEdgesPerNode = 2 //au moins deux pour s'assurer qu'un node n'est pas isolé, probabilité de configuration de trois noeuds en triangle negligé :P
	maxEdgesPerNode = 5
	weightRange     = 20 //poids max des edges
)

var numWorkers = runtime.NumCPU()
var waitGroup sync.WaitGroup

// ****		FONCTION CRÉATION GRAPHE ALÉATOIRE ****//
func initRandomGraph(nodesCount int) Graph {
	rand.Seed(time.Now().UnixNano())
	//permet d'obtenir une séquence aléatoire differente à chaque execution du code,
	//on se base sur le temps qui est un parametre qui change constantement

	// On appelle les nodes R + num comme ça il y a pas de confusion avec les poids et on a inf possibilités
	nodes := make([]*Node, nodesCount)
	for i := 0; i < nodesCount; i++ {
		nodes[i] = &Node{Name: fmt.Sprintf("R%d", i+1)}
	}

	// Creation liens aléatoirement
	for _, node := range nodes {
		// Determiner aléatoirement la quantité d'Edges que le node aura (n entre minEdgesPerNode et maxEdgesPerNode)
		edgesCount := rand.Intn(maxEdgesPerNode-minEdgesPerNode+1) + minEdgesPerNode

		for j := 0; j < edgesCount; j++ {
			// Choisir un node aléatoire
			otherNode := nodes[rand.Intn(nodesCount)]

			// Tester si le lien existe déjà et éviter un lien avec lui-même
			for node == otherNode || edgeExists(node, otherNode) {
				otherNode = nodes[rand.Intn(nodesCount)]
			}

			// Creer le lien dans les deux sens
			edge := &Edge{To: otherNode, Weight: rand.Intn(weightRange)}
			node.Edges = append(node.Edges, edge)
			otherNode.Edges = append(otherNode.Edges, &Edge{To: node, Weight: edge.Weight})
		}
	}

	return Graph{Nodes: nodes}
}

// Fct qui verifie si le lien existe déjà
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

func processMessages(g *Graph, node *Node) { //je rajoute graph pour appeler la fct qui recalcule dijkstra
	for {
		select {
		case message := <-node.Channel:
			fmt.Printf("Le nœud %s a reçu le message '%s' du nœud %s\n", message.Source.Name, message.Content, message.Destination.Name)

			// Actions selon le message reçu
			switch message.Content {
			case "Hello":
				routing(node, message)
			case "Hello Ack":
				routing(node, message)
			case "link no longer available":
				removeLinkAndRecalculate(g, message.LinkDetails) //fonction qui va enlever le lien et recalculer la routing table de tous les routeurs
			case "new link available":
				addLinkAndRecalculate(g, message.LinkDetails)
			default:
				fmt.Printf("Message de type inconnu: %s\n", message.Content)
			}
		}
	}
}

func routing(node *Node, received Message) {
	if received.Destination == node && received.Content == "Hello" {
		helloAckMessage := Message{Source: received.Destination, Destination: received.Source, Content: "Hello Ack"}
		node.RoutingTable[received.Source.Name]["next_hop"].Channel <- helloAckMessage
	} else if received.Destination != node {
		node.RoutingTable[received.Destination.Name]["next_hop"].Channel <- received
	}
}

func removeLinkAndRecalculate(g *Graph, linkinfo LinkInfo) {
	nodeA := linkinfo.NodeA
	nodeB := linkinfo.NodeB
	for i, edge := range nodeA.Edges {
		if edge.To == nodeB {
			// Eliminer le Edge de la liste de edges de A avec une technique de slicing
			nodeA.Edges = append(nodeA.Edges[:i], nodeA.Edges[i+1:]...)
			break
		}
	}
	for i, edge := range nodeB.Edges {
		if edge.To == nodeA {
			nodeB.Edges = append(nodeB.Edges[:i], nodeB.Edges[i+1:]...)
			break
		}
	}
	constructAllRoutingTables(g)
}

func addLinkAndRecalculate(g *Graph, linkinfo LinkInfo) {
	nodeA := linkinfo.NodeA
	nodeB := linkinfo.NodeB
	// j'ai besoin du poids pour créer le nouveau Edge, je le met dans la classe LinkInfo ou je fais comment?

	linkExists := false
	for _, edge := range nodeA.Edges {
		if edge.To == nodeB {
			linkExists = true
			break
		}
	}
	if !linkExists {
		// Ajout Edge au node A
		edgeA := &Edge{To: nodeB, Weight: 1} //j'ai mis 1 par default mais il faudrait plutôt avoir le parametre
		nodeA.Edges = append(nodeA.Edges, edgeA)

		// Ajout Edge au node B
		edgeB := &Edge{To: nodeA, Weight: 1}
		nodeB.Edges = append(nodeB.Edges, edgeB)

		// Recalcule RoutingTables
		constructAllRoutingTables(g)
	} else {
		fmt.Println("Le lien existait déjà.")
	}
}

// **** 		FONCTIONS CONSTRUCTION TABLES DE ROUTAGE		****//
func Dijkstra(g *Graph, start *Node) {
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
	start.RoutingTable = make(map[string]map[string]*Node)
	for destNode := range distances {
		start.RoutingTable[destNode.Name] = make(map[string]*Node)
		start.RoutingTable[destNode.Name]["next_hop"] = next_hop[destNode]
	}
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

func constructRoutingTablesWorker(jobs <-chan *Node, graph *Graph) {
	for node := range jobs {
		Dijkstra(graph, node)
		waitGroup.Done()
	}
}

func constructAllRoutingTables(graph *Graph) {
	start := time.Now()

	// Crear un canal para asignar trabajos a goroutines
	jobs := make(chan *Node, len(graph.Nodes))

	// Iniciar goroutines para construir tablas de enrutamiento
	for i := 0; i < numWorkers; i++ {
		go constructRoutingTablesWorker(jobs, graph)
	}

	// Asignar trabajos a las goroutines
	for _, node := range graph.Nodes {
		waitGroup.Add(1)
		jobs <- node
	}
	close(jobs)
	// Esperar a que todas las goroutines completen
	waitGroup.Wait()

	fmt.Printf("Tables de routage créés en %v\n", time.Since(start))
}

// **** 		FONCTION MAIN		 ****/
func main() {
	graph := initRandomGraph(1000)
	fmt.Print(numWorkers, "\n")
	constructAllRoutingTables(&graph)

	/* 	// Affichage table de routage pour chaque noeud
	   	for _, start := range graph.Nodes {
	   		fmt.Println("\nDistances les plus courtes du noeud", start.Name)
	   		for dest, route := range start.RoutingTable {
	   			fmt.Print(start.Name, " -> ", dest, " : ", route["next_hop"].Name, "\n")
	   		}
	   	} */

}
