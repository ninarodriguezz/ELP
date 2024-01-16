package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

//**** STRUCTURES ****//

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

//**** INITIALISATION ****//

// Définition des caracteristiques du graphe
const (
	minEdgesPerNode = 2 //au moins deux pour s'assurer qu'un node n'est pas isolé, probabilité de configuration de trois noeuds en triangle negligé :P
	weightRange     = 20 //poids max des edges
)

// Variables globales //
var numWorkers = runtime.NumCPU()
var waitGroup sync.WaitGroup
var dijWaitGroup sync.WaitGroup
var helloWG sync.WaitGroup
var closeWaitGroup sync.WaitGroup
var ackReceived = 0
var nodesCount int
var maxEdges int 
// **** CRÉATION GRAPHE ALÉATOIRE ****//

func initRandomGraph(nodesCount int, maxEdgesPerNode int) Graph {
	/*Docstring*/

	rand.Seed(time.Now().UnixNano())
	//permet d'obtenir une séquence aléatoire differente à chaque execution du code,
	//on se base sur le temps qui est un parametre qui change constantement

	// On appelle les nodes R + num comme ça il y a pas de confusion avec les poids et on a inf possibilités
	nodes := make([]*Node, nodesCount)

	for i := 0; i < nodesCount; i++ {
		channel := make(chan Message)
		nodes[i] = &Node{Name: fmt.Sprintf("R%d", i+1), Channel: channel}
	}

	// Creation liens aléatoirement
	for _, node := range nodes {
		// Determiner aléatoirement la quantité d'Edges que le node aura (n entre minEdgesPerNode et maxEdgesPerNode)
		edgesCount := rand.Intn(maxEdgesPerNode-minEdgesPerNode+1) + minEdgesPerNode
		if len(node.Edges) >= edgesCount {
			fmt.Print(len(node.Edges))
			for j := 0; j < edgesCount; j++ {
				// Choisir un node aléatoire
				otherNode := nodes[rand.Intn(nodesCount)]

				// Tester si le lien existe déjà et éviter un lien avec lui-même
				for node == otherNode || edgeExists(node, otherNode) || len(otherNode.Edges)>= maxEdgesPerNode{
					otherNode = nodes[rand.Intn(nodesCount)]
				}

				// Creer le lien dans les deux sens
				edge := &Edge{To: otherNode, Weight: rand.Intn(weightRange) + 1}
				node.Edges = append(node.Edges, edge)
				otherNode.Edges = append(otherNode.Edges, &Edge{To: node, Weight: edge.Weight})
			}
		}
	}

	return Graph{Nodes: nodes}
}

// Fct qui verifie si le lien existe déjà
func edgeExists(nodeA, nodeB *Node) bool {
	/*Docstring*/
	for _, edge := range nodeA.Edges {
		if edge.To == nodeB {
			return true
		}
	}
	return false
}

//**** TRANSMITION DE MESSAGES	****//

func sendMessage(messageChan chan Message, messageEnvoye Message) {
	/*Docstring*/
	messageChan <- messageEnvoye
}

func hello(nodeSrc *Node, nodeDst *Node) {
	/*Docstring*/
	channel := nodeSrc.RoutingTable[nodeDst.Name]["next_hop"].Channel
	helloMessage := Message{Source: nodeSrc, Destination: nodeDst, Content: "Hello"}
	sendMessage(channel, helloMessage)
	// fmt.Print("Message Hello envoyé depuis ", nodeSrc.Name, " à destination de ", nodeDst.Name, "\n")
	helloWG.Done()
}

func processMessages(g *Graph, node *Node) { //je rajoute graph pour appeler la fct qui recalcule dijkstra
	/*Docstring*/
	for {
		select {
		case message := <-node.Channel:
			// fmt.Printf("Le nœud %s a reçu le message '%s' destiné au nœud %s\n", node.Name, message.Content, message.Destination.Name)

			// Actions selon le message reçu
			switch message.Content {
			case "Hello":
				go routing(node, message)
			case "Hello Ack":
				go routing(node, message)
			case "link no longer available":
				waitGroup.Done()
				removeLinkAndRecalculate(g, message.LinkDetails) //fonction qui va enlever le lien et recalculer la routing table de tous les routeurs
			case "new link available":
				waitGroup.Done()
				addLinkAndRecalculate(g, message.LinkDetails)
			default:
				fmt.Printf("Message de type inconnu: %s\n", message.Content)
			}

		}
	}
}

func routing(node *Node, received Message) {
	/*Docstring*/
	if received.Destination == node && received.Content == "Hello" {
		// fmt.Print("Hello reçu par ", node.Name, " de la part de ", received.Source.Name, "\n")
		helloAckMessage := Message{Source: received.Destination, Destination: received.Source, Content: "Hello Ack"}
		nodeDst := node.RoutingTable[received.Source.Name]["next_hop"]
		sendMessage(nodeDst.Channel, helloAckMessage)
		// fmt.Print("helloAck envoyé depuis ", node.Name, " vers ", received.Source.Name, "\n")

	} else if received.Destination == node && received.Content == "Hello Ack" {
		fmt.Print(node.Name, " a reçu un message 'Hello Ack' : liaison établie entre les noeuds ", node.Name, " et ", received.Source.Name, "\n")
		ackReceived++
	} else if received.Destination != node {
		nodeDst := node.RoutingTable[received.Destination.Name]["next_hop"]
		sendMessage(nodeDst.Channel, received)
	}
}

func removeLinkAndRecalculate(g *Graph, linkinfo LinkInfo) {
	/*Docstring*/
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
	waitGroup.Done()
}

func addLinkAndRecalculate(g *Graph, linkinfo LinkInfo) {
	/*Docstring*/
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
		fmt.Print("Le lien existait déjà.\n")
	}
	waitGroup.Done()
}

// **** 		FONCTIONS CONSTRUCTION TABLES DE ROUTAGE		****//

func Dijkstra(g *Graph, start *Node) {
	/*Docstring*/
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
	/*Docstring*/
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
	/*Docstring*/
	for node := range jobs {
		Dijkstra(graph, node)
		dijWaitGroup.Done()
	}
}

func constructAllRoutingTables(graph *Graph) {
	/*Docstring*/
	start := time.Now()

	// Crear un canal para asignar trabajos a goroutines
	jobs := make(chan *Node, nodesCount)

	// Iniciar goroutines para construir tablas de enrutamiento
	for i := 0; i < numWorkers; i++ {
		go constructRoutingTablesWorker(jobs, graph)
	}

	// Asignar trabajos a las goroutines
	for _, node := range graph.Nodes {
		dijWaitGroup.Add(1)
		jobs <- node
	}

	close(jobs)

	// Esperar a que todas las goroutines completen
	dijWaitGroup.Wait()
	fmt.Printf("\nTables de routage créés en %v\n\n", time.Since(start))
}

//****		FERMETURE DE TOUS LES CHANNELS		****//

func closeChan(g Graph) {
	/*Docstring*/
	for nodeNum := 0; nodeNum < nodesCount; nodeNum++ {
		closeWaitGroup.Add(1)
		go func(node *Node) {
			defer closeWaitGroup.Done()
			close(node.Channel)
		}(g.Nodes[nodeNum])
	}

	closeWaitGroup.Wait()
}

// **** 		FONCTION MAIN		 ****/

func main() {
	/*Docstring*/

	//Création du graphe et des tables de routage pour chaque noeud
	fmt.Print("Quelle est la taille n du graphe ? (minimum n = 10) \nn = ")
	_, err := fmt.Scanln(&nodesCount)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
	if nodesCount <= 0 {
		fmt.Println("Invalid input. Size 'n' should be an integer bigger than 10.")
		return
	}
	fmt.Print("Combien d'interfaces a chaque routeur ? (minimum i = 2) \ni = ")
	_, err = fmt.Scanln(&maxEdges)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
	if nodesCount <= 0 {
		fmt.Println("Invalid input. Size 'n' should be an integer bigger than 2.")
		return
	}
	graph := initRandomGraph(nodesCount, maxEdges)
	fmt.Print(numWorkers, " CPU\n")
	constructAllRoutingTables(&graph)

	// Affichage table de routage pour chaque noeud
	// for _, start := range graph.Nodes {
	// 	fmt.Println("\nDistances les plus courtes du noeud", start.Name)
	// 	for dest, route := range start.RoutingTable {
	// 		fmt.Print(start.Name, " -> ", dest, " : ", route["next_hop"].Name, "\n")
	// 	}
	// }

	//Lancement des goroutines sur chaque noeud pour process les messages reçu
	//et envoyer un Hello Message à un noeud aléatoire
	for nodeNumber := 0; nodeNumber < nodesCount; nodeNumber++ {
		helloWG.Add(1) //Incrémentation du wait group pour la goroutine hello

		nodeSrc := graph.Nodes[nodeNumber]
		go processMessages(&graph, nodeSrc)

		nodeDst := graph.Nodes[rand.Intn(nodesCount)]
		for nodeDst == nodeSrc {
			nodeDst = graph.Nodes[rand.Intn(nodesCount)]
		}
		go hello(nodeSrc, nodeDst)

	}
	helloWG.Add(1)
	//Incrémentation du wait group pour toutes les goroutines processMessages
	//On attend que tous les noeuds aient reçu le message Hello Ack pour décrémenter le wait group
	for ackReceived < nodesCount {
	}
	helloWG.Done()
	//Se décrémente quand ackReceived s'est incrémenté jusqu'à atteindre nodesCount,
	//soit quand tous les messages Hello et Hello Ack ont fini d'être routés
	helloWG.Wait()

	//Boucle infinie pour que l'utilisateur puisse agir sur le graphe:
	//ajout ou suppression de liens, fermeture de tous les canaux
	for {

		var commande int
		fmt.Print("\nPour ajouter un lien au graphe, entrer 1.\nPour supprimer un lien existant, entrer 2.\nPour initier du traffic dans le grapghe actuel, entrer 3.\nPour fermer tous les canaux de communication, entrer 4.\nCommande 1, 2, 3 ou 4 : ")
		fmt.Scanln(&commande)

		if commande == 1 {
			//Ajout d'un lien
			var num1, num2 int
			fmt.Printf("\n\n\nVeuillez saisir un numéro de routeur : \nR")
			fmt.Scanln(&num1)
			for num1 < 1 || num1 > nodesCount {
				fmt.Printf("Saisie non valide.\nVeuillez saisir un numéro de routeur : \nR")
				fmt.Scanln(&num1)
			}
			fmt.Printf("\nVoici les voisins du routeur choisi :\n- ")
			nodeA := graph.Nodes[num1-1]

			for _, edge := range nodeA.Edges {
				fmt.Print(edge.To.Name, " - ")
			}
			fmt.Printf("\n\nVeuillez choisir le numéro d'un routeur voisin de %s :\nR", nodeA.Name)
			fmt.Scanln(&num2)
			for num2 < 1 || num2 > nodesCount {
				fmt.Printf("Saisie non valide.\nVeuillez saisir un numéro de routeur : \nR")
				fmt.Scanln(&num2)
			}
			nodeB := graph.Nodes[num2-1]

			link_details := LinkInfo{NodeA: nodeA, NodeB: nodeB}
			link_creation := Message{Source: nodeA, Destination: graph.Nodes[nodesCount-1], Content: "new link available", LinkDetails: link_details}
			waitGroup.Add(2)
			go sendMessage(graph.Nodes[nodesCount-1].Channel, link_creation)
			go processMessages(&graph, graph.Nodes[nodesCount-1])
			waitGroup.Wait()

		} else if commande == 2 {
			//Suppression d'un lien
			var num1, num2 int
			fmt.Printf("\n\n\nVeuillez saisir un numéro de routeur : \nR")
			fmt.Scanln(&num1)
			for num1 < 1 || num1 > nodesCount {
				fmt.Printf("Saisie non valide.\nVeuillez saisir un numéro de routeur : \nR")
				fmt.Scanln(&num1)
			}
			fmt.Printf("\nVoici les voisins du routeur choisi :\n- ")
			nodeA := graph.Nodes[num1-1]

			for _, edge := range nodeA.Edges {
				fmt.Print(edge.To.Name, " - ")
			}
			fmt.Printf("\n\nVeuillez choisir le numéro d'un routeur voisin de %s :\nR", nodeA.Name)
			fmt.Scanln(&num2)
			for num2 < 1 || num2 > nodesCount {
				fmt.Printf("Saisie non valide.\nVeuillez saisir un numéro de routeur : \nR")
				fmt.Scanln(&num2)
			}
			nodeB := graph.Nodes[num2-1]

			link_details := LinkInfo{NodeA: nodeA, NodeB: nodeB}
			link_failure := Message{Source: nodeA, Destination: graph.Nodes[nodesCount-1], Content: "link no longer available", LinkDetails: link_details}
			waitGroup.Add(2)
			go sendMessage(graph.Nodes[nodesCount-1].Channel, link_failure)
			go processMessages(&graph, graph.Nodes[nodesCount-1])
			waitGroup.Wait()

		} else if commande == 3 {
			//Lancement des goroutines sur chaque noeud pour process les messages reçu
			//et envoyer un Hello Message à un noeud aléatoire
			for nodeNumber := 0; nodeNumber < nodesCount; nodeNumber++ {
				helloWG.Add(1) //Incrémentation du wait group pour la goroutine hello

				nodeSrc := graph.Nodes[nodeNumber]
				go processMessages(&graph, nodeSrc)

				nodeDst := graph.Nodes[rand.Intn(nodesCount)]
				for nodeDst == nodeSrc {
					nodeDst = graph.Nodes[rand.Intn(nodesCount)]
				}
				go hello(nodeSrc, nodeDst)

			}
			helloWG.Add(1)
			//Incrémentation du wait group pour toutes les goroutines processMessages
			//On attend que tous les noeuds aient reçu le message Hello Ack pour décrémenter le wait group
			for ackReceived < nodesCount {
			}
			helloWG.Done()
			//Se décrémente quand ackReceived s'est incrémenté jusqu'à atteindre nodesCount,
			//soit quand tous les messages Hello et Hello Ack ont fini d'être routés
			helloWG.Wait()
			time.Sleep(2 * time.Second)
		} else if commande == 4 {
			break
		} else {
			fmt.Print("\nVeillez à entrer 1, 2, 3 ou 4\n")
		}
	}

	closeChan(graph)

}

/*

// Affichage table de routage pour chaque noeud
for _, start := range graph.Nodes {
	fmt.Println("\nDistances les plus courtes du noeud", start.Name)
	for dest, route := range start.RoutingTable {
		fmt.Print(start.Name, " -> ", dest, " : ", route["next_hop"].Name, "\n")
	}
} */
