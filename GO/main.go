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
	Route       map[*Node]struct{}
	LinkDetails LinkInfo //info contenue dans les messages pour informer qu'on a perdu ou établi un nouveau lien
}

type LinkInfo struct {
	NodeA *Node
	NodeB *Node //nodes qui ont perdu ou récuperé un lien
}

//**** INITIALISATION ****//

// Définition des caracteristiques du graphe
const (
	minEdgesPerNode = 2  //au moins deux pour s'assurer qu'un node n'est pas isolé, probabilité de configuration de trois noeuds en triangle negligé :P
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
	/*
		initRandomGraph initialise et retourne un graphe aléatoire caractérisé par les paramètres de la fonction.

		Paramètres :
			- nodesCount : le nombre de nœuds dans le graphe
			- maxEdgesPerNode : le nombre maximal d'arêtes par nœud

		La fonctioneffectue un tirage aléatoire basé sur le temps pour garantir une séquence
		aléatoire différente à chaque exécution du code. Les nœuds du graphe sont créés avec des
		canaux de messages associés et des noms distincts (R + numéro). Les liens entre les nœuds
		sont établis de manière aléatoire, en évitant les doublons et les liens avec eux-mêmes (arête boucle).

		Retourne :
			- Un objet Graph représentant le graphe initialisé

	*/

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
		if len(node.Edges) < edgesCount {
			for j := len(node.Edges); j < edgesCount; j++ {
				// Choisir un node aléatoire
				otherNode := nodes[rand.Intn(nodesCount)]

				count := 0 // Counter pour éviter une boucle infinie, si on ne trouve pas un node disponible après n/2 essais, on arrete les random

				// Tester si le lien existe déjà et éviter un lien avec lui-même ou avec un routeur qui a toutes les intérfaces occupées
				for node == otherNode || edgeExists(node, otherNode) || len(otherNode.Edges) >= maxEdgesPerNode {
					otherNode = nodes[rand.Intn(nodesCount)]
					count += 1
					if count > nodesCount/2 && len(node.Edges) >= minEdgesPerNode { //quand on a déjà essayé n/2 fois on arrête de chercher un noeud random
						break
					} else if count > nodesCount/2 && len(node.Edges) < minEdgesPerNode {
						for _, otherNode = range nodes {
							if len(otherNode.Edges) > minEdgesPerNode+1 {
								otherNode.Edges[0].To = node
								break
							}
						}

					}

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

func edgeExists(nodeA, nodeB *Node) bool {
	/*
		edgeExists verifie si le lien défini par les noeuds en paramètre existe déjà

		Paramètres :
			- nodeA : noeud à une extrémité du lien
			- nodeB : noeud à l'autre extrémité

		Retourne :
			- Un booléen true si le lien existe, false sinon
	*/
	for _, edge := range nodeA.Edges {
		if edge.To == nodeB {
			return true
		}
	}
	return false
}

//**** TRANSMITION DE MESSAGES	****//

func sendMessage(messageChan chan Message, messageEnvoye Message) {
	/*
		sendMessage envoie un message dans un canal de communication.

		Paramètres :
			- messageChan : canal de communication dans lequel est envoyé le message
			- messageEnvoyé : message à envoyer dans le canal de communication

		La fonction ne retourne rien.
	*/
	messageChan <- messageEnvoye
}

func hello(nodeSrc *Node, nodeDst *Node) {
	/*
		hello envoie un message de type "Hello" du nœud source au nœud destination.

		Paramètres :
			- nodeSrc : Le nœud source à partir duquel le message "Hello" est envoyé.
			- nodeDst : Le nœud destination auquel le message "Hello" est adressé.

		La fonction utilise la table de routage du nœud source pour déterminer le canal de communication
		du noeud correspondant au prochain saut vers le nœud destination. Elle crée ensuite un message de type
		"Hello" avec le nœud source comme émetteur et le nœud destination comme destinataire, puis envoie
		ce message sur le canal spécifié.

		Une fois l'envoi terminé, helloWG.Done() est appelé pour décrémenter le compteur du WaitGroup helloWG.

		La fonction ne retourne rien.
	*/
	channel := nodeSrc.RoutingTable[nodeDst.Name]["next_hop"].Channel
	route := make(map[*Node]struct{})
	route[nodeSrc] = struct{}{}
	helloMessage := Message{Source: nodeSrc, Destination: nodeDst, Content: "Hello", Route: route}
	sendMessage(channel, helloMessage)
	// fmt.Print("Message Hello envoyé depuis ", nodeSrc.Name, " à destination de ", nodeDst.Name, "\n")
	helloWG.Done()
}

func processMessages(g *Graph, node *Node) {
	/*
		processMessages écoute en permanence les messages provenant du canal du nœud spécifié en paramètre
		et effectue des actions en conséquence selon du contenu du message.

		Paramètres :

			- g : Le graphe global contenant l'ensemble des nœuds
			- node : Le nœud actuel pour lequel les messages sont traités

		La fonction utilise une boucle infinie pour écouter les messages du canal du nœud en permanence.
		Lorsqu'un message est reçu, la fonction effectue des actions dépendantes du type de message reçu.
		La fonction prend en charge les messages de type "Hello", "Hello Ack", "link no longer available",
		et "new link available". Pour chaque type de message, la fonction fait appel des fonctions spécifiques
		pour traiter le message.

		La fonction ne retourne rien.
	*/
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
			}

		}
	}
}

func routing(node *Node, received Message) {
	/*
			 	routing traite le message reçu (de type Hello ou Hello Ack) en fonction du nœud actuel et
				du contenu du message.

		 		Paramètres :
		   			- node : Le nœud actuel qui traite le message
		   			- received : Le message reçu à traiter

		 		La fonction examine le contenu du message et le traite selon sa destination et son contenu.
				Si le message est de type "Hello" et est destiné au nœud actuel, un message "Hello Ack" est
				envoyé à la source du message initial. Si le message est de type "Hello Ack" et est destiné
				au nœud actuel, un message est affiché indiquant l'établissement de la liaison entre les nœuds.
				Pour un message (peu importe son type) qui n'est pas destiné au noeud actuel, le message est
				transmis au prochain saut déterminé par la table de routage.
	*/

	received.Route[node] = struct{}{}

	if received.Destination == node && received.Content == "Hello" {
		// fmt.Print("Hello reçu par ", node.Name, " de la part de ", received.Source.Name, " -- Route: ", afficherRoute(received.Route), "\n")
		route := make(map[*Node]struct{})
		route[node] = struct{}{}
		helloAckMessage := Message{Source: received.Destination, Destination: received.Source, Content: "Hello Ack", Route: route}
		nodeDst := node.RoutingTable[received.Source.Name]["next_hop"]
		sendMessage(nodeDst.Channel, helloAckMessage)
		// fmt.Print("helloAck envoyé depuis ", node.Name, " vers ", received.Source.Name, "\n")

	} else if received.Destination == node && received.Content == "Hello Ack" {
		fmt.Print(node.Name, " a reçu un message 'Hello Ack' : liaison établie entre les noeuds ", node.Name, " et ", received.Source.Name, "\nRoute : ", afficherRoute(received.Route), "\n")
		ackReceived++
	} else if received.Destination != node {
		nodeDst := node.RoutingTable[received.Destination.Name]["next_hop"]
		sendMessage(nodeDst.Channel, received)
	}
}

func afficherRoute(route map[*Node]struct{}) string {
	/*
			afficherRoute crée une représentation sous forme de chaîne de caractères
			d'une route spécifiée, en utilisant les noms des nœuds dans l'ordre de la route.

			Paramètre :
		   		- route : La map représentant la route, avec les nœuds comme clés

		   	Retourne :
		   		- Une chaîne de caractères représentant la route

	*/
	var toPrint string
	for node := range route {
		toPrint += " " + node.Name + " "
	}
	return toPrint
}

func removeLinkAndRecalculate(g *Graph, linkinfo LinkInfo) {
	/*
		removeLinkAndRecalculate supprime le lien entre deux nœuds dans le graphe
		et recalcule les tables de routage de tous les nœuds du graphe.

		Paramètres :
		   - g : Le graphe global contenant l'ensemble des nœuds
		   - linkinfo : Les informations sur le lien à supprimer, dont les nœuds reliés par ce lien

		La fonction recherche le lien entre nodeA et nodeB dans les listes d'arêtes des deux
		nœuds et le supprime. Ensuite, la fonction appelle la fonction constructAllRoutingTables
		pour recalculer les tables de routage de tous les nœuds du graphe, en prenant en compte
		la suppression du lien. Enfin, la fonction décrémente le compteur de WaitGroup.

		La fonction ne retourne rien.
	*/
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
	/*
		addLinkAndRecalculate ajoute un lien entre deux nœuds dans le graphe,
		et recalcule les tables de routage de tous les nœuds du graphe.

		Paramètres :
			- g : Le graphe global contenant l'ensemble des nœuds
			- linkinfo : Les informations sur le lien à ajouter, dont les noeuds reliés par ce lien

		La fonction vérifie d'abord si le lien entre nodeA et nodeB existe déjà. Si le lien
		n'existe pas, un nouvel Edge est créé pour chaque nœud et ajouté à leur liste d'arêtes.
		Ensuite, la fonction appelle la fonction constructAllRoutingTables pour recalculer les
		tables de routage de tous les nœuds du graphe, en prenant en compte l'ajout du nouveau lien.
		Enfin, la fonction décrémente le compteur du WaitGroup.

		La fonction ne retourne rien.
	*/
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
	/*
			Dijkstra applique l'algorithme de Dijkstra pour calculer les tables de routage
			à partir d'un nœud de départ dans le graphe entré en paramètre.

			Paramètres :
		   		- g : Le graphe global contenant l'ensemble des nœuds et des liens
		   		- start : Le nœud de départ à partir duquel l'algorithme de Dijkstra est lancé

			La fonction utilise l'algorithme de Dijkstra pour calculer les distances minimales
			entre le nœud de départ et tous les autres nœuds du graphe. Elle construit ensuite
			la table de routage du nœud de départ en utilisant les résultats de l'algorithme.
			Les tables de routage indiquent le prochain nœud (next_hop) sur le chemin le plus
			court vers chaque destination.

			La fonction ne retourne rien.
	*/
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
	/*
		minDist retourne le nœud non visité ayant la distance minimale dans la map de distances.

		Paramètres :
			- unvisited : Une map contenant les nœuds non visités
			- distances : Une map contenant les distances minimales actuelles pour chaque nœud

		Retourne :
			- Le nœud non visité ayant la distance minimale, ou nil si la map est vide
	*/
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
	/*
		constructRoutingTablesWorker utilise l'algorithme de Dijkstra pour calculer les tables de
		routage de tous les noeuds du graphe entré en paramètre.

		Paramètres :
			- jobs : Un canal (channel) fournissant les nœuds à traiter avec l'algorithme de Dijkstra
			- graph : Le graphe global contenant l'ensemble des nœuds

		La fonction reçoit des nœuds à partir du canal "jobs" et applique l'algorithme de Dijkstra
		à chaque nœud pour calculer les tables de routage correspondantes. La goroutine décrémente
		le compteur du WaitGroup dijWaitGroup après chaque exécution de l'algorithme sur un nœud.

		La fonction ne retourne rien.
	*/
	for node := range jobs {
		Dijkstra(graph, node)
		dijWaitGroup.Done()
	}
}

func constructAllRoutingTables(graph *Graph) {
	/*
			constructAllRoutingTables crée les tables de routage de tous les nœuds dans le graphe
		 	en utilisant plusieurs goroutines pour accélérer le processus.

		 	Paramètres :
		   		- graph : Le graphe global contenant l'ensemble des nœuds

		 	La fonction utilise plusieurs goroutines pour appliquer simultanément l'algorithme de Dijkstra
			à chaque nœud du graphe, afin de calculer les tables de routage correspondantes. Elle utilise un
		 	canal pour assigner des travaux aux goroutines, puis attend que toutes les goroutines aient
		 	terminé leur travail avant de fournir le temps écoulé depuis le début de la création des tables
		 	de routage.

			Le fonction ne retourne rien.
	*/
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
	/*
		closeChan ferme les canaux de communication associés à chaque nœud dans le graphe.

		Paramètres :
			- g : Le graphe global contenant l'ensemble des nœuds

		La fonction itère sur tous les nœuds du graphe et ferme leur canal de communication,
		en utilisant des goroutines pour accélérer le processus. Elle utilise un WaitGroup closeWaitGroup
		pour attendre que toutes les goroutines aient terminé avant de retourner.
	*/
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
	/*
		main est la fonction principale du programme qui simule un réseau de routeurs
		et teste la communication entre eux en utilisant l'algorithme de Dijkstra pour
		construire les tables de routage.

		L'utilisateur est invité à définir la taille du graphe, le nombre d'interfaces
		par routeur, et peut ensuite effectuer différentes actions telles que l'ajout ou
		la suppression de liens, l'initiation de trafic entre tous les routeurs,
		l'initialisation de trafic entre deux routeurs au choix ou encore
		la fermeture de tous les canaux de communication.
	*/

	//Création du graphe et des tables de routage pour chaque noeud
	fmt.Print("Quelle est la taille N du graphe ? (minimum N = 10) \nN = ")
	_, err := fmt.Scanln(&nodesCount)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
	if nodesCount < 10 {
		fmt.Println("Invalid input. N doit être un entier supérieur à 10.")
		return
	}
	fmt.Print("Combien d'interfaces a chaque routeur ? (minimum i = 3) \ni = ")
	_, err = fmt.Scanln(&maxEdges)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
	if maxEdges < 2 {
		fmt.Println("Invalid input. 'i' doit être supérieur à 2.")
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
		fmt.Print("\n1 - Pour ajouter un lien au graphe.\n2 - Pour supprimer un lien existant.\n3 - Pour initier du traffic dans le graphe actuel.\n4 - Pour initier du traffic entre deux routeurs.\n5 - Pour fermer tous les canaux de communication.\nCommande 1, 2, 3, 4 ou 5 : ")
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
			fmt.Printf("\n\nVeuillez choisir le numéro d'un routeur qui n'est pas voisin de %s :\nR", nodeA.Name)
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
			for num1 < 1 || num1 > nodesCount || len(graph.Nodes[num1-1].Edges) == maxEdges {
				fmt.Printf("Saisie non valide. Le routeur n'existe pas ou ses interfaces sont déjà complètes.\nVeuillez saisir un numéro de routeur : \nR")
				fmt.Scanln(&num1)
			}
			fmt.Printf("\nVoici les voisins du routeur choisi :\n- ")
			nodeA := graph.Nodes[num1-1]

			for _, edge := range nodeA.Edges {
				fmt.Print(edge.To.Name, " - ")
			}
			fmt.Printf("\n\nVeuillez choisir le numéro d'un routeur voisin de %s :\nR", nodeA.Name)
			fmt.Scanln(&num2)
			for num2 < 1 || num2 > nodesCount || len(graph.Nodes[num1-1].Edges) == maxEdges {
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
			var num1, num2 int
			fmt.Printf("\n\n\nVeuillez saisir un numéro de routeur : \nR")
			fmt.Scanln(&num1)
			for num1 < 1 || num1 > nodesCount {
				fmt.Printf("Saisie non valide.\nVeuillez saisir un numéro de routeur : \nR")
				fmt.Scanln(&num1)
			}
			nodeA := graph.Nodes[num1-1]

			fmt.Printf("\n\nVeuillez choisir le numéro d'un autre routeur :\nR")
			fmt.Scanln(&num2)
			for num2 < 1 || num2 > nodesCount || num2 == num1 {
				fmt.Printf("Saisie non valide.\nVeuillez saisir un numéro de routeur : \nR")
				fmt.Scanln(&num2)
			}
			nodeB := graph.Nodes[num2-1]

			for nodeNumber := 0; nodeNumber < nodesCount; nodeNumber++ {

				nodeSrc := graph.Nodes[nodeNumber]
				go processMessages(&graph, nodeSrc)

			}
			helloWG.Add(2)
			go hello(nodeA, nodeB)
			//Incrémentation du wait group pour toutes les goroutines processMessages
			//On attend que tous les noeuds aient reçu le message Hello Ack pour décrémenter le wait group
			for ackReceived < 1 {
			}
			helloWG.Done()
			//Se décrémente quand ackReceived s'est incrémenté jusqu'à atteindre nodesCount,
			//soit quand tous les messages Hello et Hello Ack ont fini d'être routés
			helloWG.Wait()
			time.Sleep(2 * time.Second)

		} else if commande == 5 {
			break
		} else {
			var dummyInt int
			var dummyStr string   // Variable pour vider le buffer
			fmt.Scanln(&dummyInt) // On lit s'il reste quelque chose dans le buffer
			fmt.Scanln(&dummyStr)
			fmt.Print("\nSaisie incorrecte.\nVeillez à entrer 1, 2, 3 ou 4\n")

		}
	}
	closeChan(graph)

}
