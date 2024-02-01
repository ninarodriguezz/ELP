var prompt = require('prompt');

prompt.start();

let players = [];

function onErr(err) {
    console.log(err);
    return 1;
}

function getPlayerName(playerNumber) {
    prompt.get([{
        name: 'playerName',
        description: `Player ${playerNumber}, what is your name?`,
        type: 'string',
        required: true
    }], function (err, result) {
        if (err) { return onErr(err); }
        let player = {
            name: result.playerName,
            letters: [],
            words: [],
        };
        players.push(player);
        console.log(`Player ${result.playerName} has joined the game`);
        if (players.length < 2) {
            getPlayerName(2);
        } else {
            // Both players have joined, start the game
            setupGame(players.map(p => p.name));
        }
    });
}

// Start by getting the first player's name
getPlayerName(1);

// Define the pool of letters
const letters = 'A'.repeat(14) + 'B'.repeat(4) + 'C'.repeat(7) + 'D'.repeat(5) + 'E'.repeat(19) + 'F'.repeat(2) + 'G'.repeat(4) + 'H'.repeat(2) + 'I'.repeat(11) + 'J'.repeat(1) + 'K'.repeat(1) + 'L'.repeat(6) + 'M'.repeat(5) + 'N'.repeat(9) + 'O'.repeat(8) + 'P'.repeat(4) + 'Q'.repeat(1) + 'R'.repeat(10) + 'S'.repeat(7) + 'T'.repeat(9) + 'U'.repeat(8) + 'V'.repeat(2) + 'W'.repeat(1) + 'X'.repeat(1) + 'Y'.repeat(1) + 'Z'.repeat(2);
const lettersArray = letters.split('');

// Function to draw a random letter from the pool
function drawLetter() {
    const index = Math.floor(Math.random() * lettersArray.length);
    return lettersArray.splice(index, 1)[0];
}

// Function to setup the game
function setupGame(playerNames) {
    // Reset the game state
    let gameState = {
        players: [],
        currentplayer: 0,
    };

    // Add each player to the game state
    for (const playerName of playerNames) {
        const player = {
            name: playerName,
            score: 0,
            letters: [],
            words: [],
            move: null,  //voir si c'est nécessaire ou pas	
        };

        // Draw 6 letters for the player
        for (let i = 0; i < 6; i++) {
            player.letters.push(drawLetter());
        }

        gameState.players.push(player);
    }
    startGame(gameState);

}

// Function to start the game
function startGame(gameState) {
    displayGameState(gameState);
}

function displayGameState(gameState) {
    // Display each player's board, letters, and points
    for (const player of gameState.players) {
        console.log(`${player.name}'s board:`);
        for (const word of player.words) {
            console.log(word.split('').join(' '));
        }
        console.log(`${player.name}'s letters: ${player.letters.join(' ')}`);
        console.log(`${player.name}'s points: ${player.score}`);
    }
}


// Function to handle a player making a move
function makeMove(playerName, move) {
    // Find the player in the game state
    const player = gameState.players.find((p) => p.name === playerName);

    // Update the player's move
    player.move = move;

    // Return the updated game state
    return gameState;
}

// Function to get the current game state
function getGameState() {
    return gameState;
}

// Export the functions for use in other modules
module.exports = {
    makeMove,
    getGameState,
    setupGame,
    startGame,
    addWord,
};


function addWord(playerName, word) {
    // Find the player
    const player = gameState.players.find(p => p.name === playerName);

    // Points for each word length
    const pointsArray = [0, 0, 9, 16, 25, 36, 49, 64, 81];

    // Calculate the points for the word
    const points = pointsArray[word.length];

    // Add the word to the player's words array
    player.words.push({ word, points });
}

/* function joinGame(playerName) {
    // Create a new player object
    const newPlayer = {
        name: playerName,
        letters: [],
        words: [], // Array to store the words the player has formed
    };

    // Add the new player to the game state
    gameState.players.push(newPlayer);
} */
