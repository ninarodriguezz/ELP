package main

import (
	"fmt"
	"time"
)

func main() {

	var graphe [5][5]int
	graphe[0][0], graphe[0][1], graphe[0][2], graphe[0][3], graphe[0][4] = -1, -1, 2, 3, -1
	graphe[1][0], graphe[1][1], graphe[1][2], graphe[1][3], graphe[1][4] = -1, -1, 2, 5, -1
	graphe[2][0], graphe[2][1], graphe[2][2], graphe[2][3], graphe[2][4] = 1, 1, -1, -1, -1
	graphe[3][0], graphe[3][1], graphe[3][2], graphe[3][3], graphe[3][4] = 1, -1, -1, -1, 7
	graphe[4][0], graphe[4][1], graphe[4][2], graphe[4][3], graphe[4][4] = 1, 6, 6, 2, -1

	for sommet := 0; sommet < len(graphe); sommet++ {
		go dijkstra(graphe, sommet)
		time.Sleep(time.Second * 3)
	}

	fmt.Print("fin\n")

}

func isIn(a int, b []int) bool {
	for i := 0; i < len(b); i++ {
		if b[i] == a {
			return true
		}
	}
	return false
}

func trouve_min(d [5]int, marqués []int) int {

	min := 100000
	sommet := -1

	for i := 0; i < len(d); i++ {

		if !isIn(i, marqués) && d[i] < min {
			min = d[i]
			sommet = i
		}
	}
	return sommet
}

func dijkstra(G [5][5]int, s int) {

	// Initialisation
	var d [5]int
	var marqués []int
	for i := 0; i < len(G); i++ {
		d[i] = 10000
	}
	d[s] = 0

	// Itération sur les sommets non marqués
	for len(marqués) != len(G) {

		// On marque le sommet le plus proche de sommet de départ
		a := trouve_min(d, marqués)
		marqués = append(marqués, a)

		// Itération sur les sommets voisins (non marqués) du sommet courant
		for b := 0; b < len(G); b++ {

			if !isIn(b, marqués) && G[a][b] != -1 {

				// On compare la distance du voisin au sommet actuelle du sommet voisin au sommet de départ avec sa distance en passant par le sommet courant
				if d[b] > d[a]+G[a][b] {
					d[b] = d[a] + G[a][b]
				}
			}

		}
	}

	fmt.Print("Sommet source = ", s, " ; Vecteur de distance : ", d, "\n")
}
