***Description***
Le programme implémente un simulateur de réseau basé sur des routeurs et des liaisons entre eux. Il utilise des goroutines pour simuler le traitement asynchrone des messages entre les routeurs. Chaque routeur a sa propre table de routage, calculée à l'aide de l'algorithme de Dijkstra. Les messages de type "Hello" et "Hello Ack" sont échangés entre les routeurs pour établir des connexions entre les routeurs éloignés. 

***Structures principales*** 
- Graph 
Contient le "array" des Nodes du graph 

- Node 
Représente un sommet du graphe. Il contient le nom du sommet, ses liens vers d'autres sommets, son canal de communication avec lequel il reçoit des messages et sa table de routage. 

- Edge
Représente la liason entre deux sommets du graphe. Il est defini de manière unidirectionnelle grâce à l'attribut "To" mais on s'assure que tous les liens du graphe soient bidirectionnels. On définit aussi le poids du lien. 

- Message 
Contient les sommets source et destination, le contenu texte du message, la route qu'il a empruntée et, éventuellement, les details du lien à modifier. 

- LinkInfo 
Contient les deux sommets du lien à modifier. 


***Structure et Fonctionnalités*** 
Le code est structuré en plusieurs parties, notamment l'initialisation du graphe, le calcul des tables de routage, la transmission de messages et le choix de l'utilisateur pour faire des modifications dans le réseau.

- Initialisation du Graphe:

Le graphe est initialisé avec un nombre spécifié de routeurs.
Chaque routeur a un nombre défini d'interfaces (liaisons) avec d'autres routeurs. Ces valeurs sont choisies par l'utilisateur avec quelques restrictions (min 10 routers et 3 interfaces par routeur).

- Construction des Tables de Routage:

Les tables de routage de chaque routeur sont construites à l'aide de l'algorithme de Dijkstra.
Les distances minimales et les prochains sauts vers chaque destination sont calculés.

- Échange de Messages:

Les routeurs échangent des messages de type "Hello" et "Hello Ack" pour établir des liaisons.
Les messages "link no longer available" et "new link available" sont utilisés pour signaler la suppression ou l'ajout de liaisons.

- Simulation du Trafic:

Le programme simule le trafic en lançant des messages "Hello" depuis tous les routeurs vers d'autres routeurs aléatoires de manière asynchrone.
Les routeurs échangent également des "Hello Ack" pour confirmer l'établissement de liaisons.
Il est ultérieurement possible pour l'utilisateur d'initier du trafic entre deux routeurs de son choix. Le premier routeur choisi va lancer un message "Hello" à destination du second routeur qui, à la réception de ce "Hello", va alors envoyer un "Hello Ack" vers le premier routeur.

- Modification Dynamique du Graphe:

L'utilisateur peut ajouter ou supprimer des liaisons entre les routeurs pendant l'exécution du programme.
Les tables de routage sont mises à jour en conséquence.

- Fermeture des Canaux:

L'utilisateur peut fermer tous les canaux de communication entre les routeurs et arrêter le programme. 

***Instructions d'Exécution***

Exécutez le programme en utilisant un environnement Go avec la commande go run main.go.
Suivez les instructions pour spécifier la taille du graphe et le nombre d'interfaces par routeur.
Le programme peut afficher les tables de routage initiales (le code de cet affichage est actuellement commenté en prévision de grands graphes) et lance la simulation du trafic.
L'utilisateur peut entrer des commandes pour ajouter ou supprimer des liaisons, initier du trafic ou fermer tous les canaux de communication entre les routeurs.
Pour ajouter ou supprimer des liaisons, suivez les instructions et saisissez les numéros des routeurs concernés. Idem pour initier du trafic entre deux routeurs au choix.

**Exemple d'Utilisation**

Initialiser un graphe avec 100 routeurs et 5 interfaces par routeur.
Possibilité d'affichage des tables de routage initiales (le code est commenté dans le main).
Lancer la simulation du trafic entre deux routeurs au choix ou entre tous les routeurs aléatoirement avec des messages "Hello".
Ajouter ou supprimer des liaisons entre les routeurs.
Possibilité de voir les changements de route quand on envoie des "Hello" entre deux routeurs. 
Fermer tous les canaux de communication pour terminer le programme.
