package main

import (
	"fmt"
	"math/rand"
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
	RoutingTable map[string]map[string]*Node  //Table de routage de chaque node qui contient tous les autres sommets avec le next_hop (sans distance)
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
	LinkDetails LinkInfo 		//info contenue dans les messages pour informer qu'on a perdu ou établi un nouveau lien 
}

type LinkInfo struct {
	NodeA *Node
	NodeB *Node 				//nodes qui ont perdu ou récuperé un lien 
}

//definition caracteristiques du graphe
const ( 
	minEdgesPerNode = 2		//au moins deux pour s'assurer qu'un node n'est pas isolé, probabilité de configuration de trois noeuds en triangle negligé :P 
	maxEdgesPerNode = 5
	weightRange     = 20	//poids max des edges
)

//****		FONCTION CRÉATION GRAPHE ALÉATOIRE ****//
func initRandomGraph(nodesCount int) Graph {
	rand.Seed(time.Now().UnixNano())			
	//permet d'obtenir une séquence aléatoire differente à chaque execution du code, 
	//on se base sur le temps qui est un parametre que change constantement 

	// On appele les nodes R + num comme ça il y a pas de confusion avec les poids et on a inf possibilités
	nodes := make([]*Node, nodesCount)
	for i := 0; i < nodesCount; i++ {
		nodes[i] = &Node{Name: fmt.Sprintf("R%d", i+1)}
	}

	// Crear conexiones aleatorias entre nodos
	for _, node := range nodes {
		// Determiner aléatoriamente la quantité d'Edges que le node aura (n entre minEdgesPerNode et maxEdgesPerNode)
		edgesCount := rand.Intn(maxEdgesPerNode-minEdgesPerNode+1) + minEdgesPerNode

		// Crear edges
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

/* func sendMessage(messageChan chan Message, messageEnvoye Message) {
	messageChan <- messageEnvoye
}

func hello(nodeSrc *Node, nodeDst *Node) {
	channel := nodeSrc.Channel
	helloMessage := Message{Source: nodeSrc, Destination: nodeDst, Content: "Hello"}
	sendMessage(channel, helloMessage)
}

func processMessages(g *Graph, node *Node) {  //je rajoute graph pour appeler la fct qui recalcule dijkstra
	for {
		select {
		case message := <-node.Channel:
			fmt.Printf("Le nœud %s a reçu le message '%s' du nœud %s\n", message.Source.Name, message.Content, message.Destination.Name)

			// Actions selon le message reçu
			switch message.Content {
			case "Hello":
				//On fait qqch si on reçoit "Hello"?? On envoie un autre message?  
			case "link no longer available":
				removeLinkAndRecalculate(g, message.LinkDetails) //fonction qui va enlever le lien et recalculer la routing table de tous les routeurs
			case "new link available":
				// addLinkAndRecalculate(g, message.LinkInfo)
			default:
				// Lógica por defecto o manejo de otros tipos de mensajes
				fmt.Printf("Message de type inconnu: %s\n", message.Content)
			}
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

func removeLinkAndRecalculate(g *Graph, linkinfo LinkInfo) {
	nodeA = linkinfo.NodeA
	nodeB = linkinfo.NodeB
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
	// il manque l'appel à dijkstra paralellisé 
}
 */
//func 

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

func constructRoutingTables(g *Graph) {
	var wg sync.WaitGroup

	for _, start := range g.Nodes {
		wg.Add(1)
		go func(start *Node) {
			defer wg.Done()
			distances, nextHop := Dijkstra(g, start)

			start.RoutingTable = make(map[string]map[string]*Node)
			for node := range distances {
				start.RoutingTable[node.Name] = make(map[string]*Node)
				start.RoutingTable[node.Name]["next_hop"] = nextHop[node]
			}
		}(start)
	}

	wg.Wait()
}


//**** 		FONCTION MAIN		 ****//

func main() {

	graph := initRandomGraph(15) //en parametre on met le nombre de sommets
	constructRoutingTables(&graph)

	// Affichage table de routage pour chaque noeud
	for _, start := range graph.Nodes {
		fmt.Println("\nDistances les plus courtes du noeud", start.Name)
		for dest, route := range start.RoutingTable {
			fmt.Print(start.Name, " -> ", dest, " : ", route["next_hop"].Name, "\n")
		}
	}
	
}