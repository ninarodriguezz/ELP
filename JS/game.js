const prompt = require('prompt');
const util = require('util');
const fs = require('fs');
const logFile = 'game_log.txt';


// Use of fs.writeFile to clean the file content
prompt.start();
const get = util.promisify(prompt.get);

let gameState = {
    players: [],
    currentPlayer: 0,
};

function onErr(err) {
    console.log(err);
    return 1;
}

function cleanFile() {
    fs.writeFile(logFile, '', (err) => {
        if (err) {
        console.error(err);
        return;
        }
    });
}    

function getPlayerName(playerNumber) {
    prompt.get([{
        name: 'playerName',
        description: `Player ${playerNumber}, what is your name?`,
        type: 'string',
        required: true
    }], function (err, result) {
        if (err) { return onErr(err); }
        const player = {
            name: result.playerName,
            score: 0,
            letters: [],
            words: [],
            move: null
        };
        gameState.players.push(player);
        console.log(`Player ${result.playerName} has joined the game`);
        if (gameState.players.length < 2) {
            getPlayerName(2);
        } else {
            // Both players have joined, start the game
            setupGame(gameState.players.map(p => p.name));
        }
    });
}

cleanFile();
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
    // Add each player to the game state
    for (const playerName of playerNames) {
        // Find the player in gameState.players
        let player = gameState.players.find(p => p.name === playerName);

        // Draw 6 letters for the player
        for (let i = 0; i < 6; i++) {
            player.letters.push(drawLetter());
        }
    }
    startGame(gameState);
}

async function startGame(gameState) {
    gameState.currentPlayer = 0;
    if (!gameState.players || !Array.isArray(gameState.players)) {
        console.error('gameState.players is not an array');
        return;
    }

    while (!checkEndCondition(gameState)) {
        if (gameState.currentPlayer === undefined || gameState.currentPlayer >= gameState.players.length) {
            console.error('gameState.currentPlayer is not a valid index');
            return;
        }

        displayGameState(gameState);
        await playerTurn(gameState.players[gameState.currentPlayer]);
        gameState.currentPlayer = (gameState.currentPlayer + 1) % gameState.players.length;
        //Ask the player if he wants to do a "jarnac"
        const jarnacResult = await get([{
            name: 'jarnac',
            description: gameState.players[gameState.currentPlayer].name + ', do you want to do a "jarnac"? (yes/no)',
            type: 'string',
            required: true
        }]);
        if (jarnacResult.jarnac.toLowerCase() === 'yes') {
            //call the function jarnac
            jarnac(gameState.currentPlayer)

        }
}
}

function displayGameState() {
    // Display each player's board, letters, and points
    try {
        for (const player of gameState.players) {
            console.log(`${player.name}'s board:`);
            for (const word of player.words) {
                if (word) {
                    console.log(word.split('').join(' '));
                }
            }
        console.log(`${player.name}'s letters: ${player.letters.join(' ')}`);
        console.log(`${player.name}'s points: ${player.score}`);
        }
    } catch (err) {
        console.error('Error displaying game state:', err);
    }
}

function checkEndCondition(gameState) {
    for (let player of gameState.players) {
        if (player.words.length === 8) {
            // The game ends if a player has 8 words
            return true;
        }
    }
    // The game continues if no player has 8 words
    return false;
}

function determineWinner(gameState) {
    let highestScore = 0;
    let winner = null;

    for (let player of gameState.players) {
        let score = player.words.reduce((total, word) => total + Math.pow(word.length, 2), 0); 
        if (score > highestScore) {
            highestScore = score;
            winner = player;
        }
    }

    return winner;
}

async function playerTurn(player) {
    let playAgain = true;

    while (playAgain) {
        // Ask the player if they want to play a word or pass their turn
        const actionResult = await get([{
            name: 'action',
            description: `${player.name}, do you want to play a word or pass your turn? (play/pass)`,
            type: 'string',
            required: true
        }]);

        if (actionResult.action.toLowerCase() === 'pass') {
            console.log(`${player.name} has decided to pass their turn.`);
            playAgain = false;
            continue;
        }

        // Ask the player for a word and the position to play it
        const result = await get([{
            name: 'word',
            description: `${player.name}, enter a word to play`,
            type: 'string',
            required: true
        }, {
            name: 'position',
            description: 'Enter the line where you want to play the word',
            type: 'number',
            required: true,
            conform: function(value) {
                const maxPosition = player.words.length + 1;
                return value >= 1 && value <= maxPosition;
            },
            message: 'Position must be between 1 and ' + (player.words.length + 1)
        }]);
        
        // Convert the word to uppercase
        const word = result.word.toUpperCase();

        // Check if the word is possible
        if (checkWord(player.letters, player.words, word, result.position)) {
            // If the word is possible, make the move
            const move = { word: word, position: result.position };
            await makeMove(player.name, move);

            // Draw a new letter for the player
            player.letters.push(drawLetter());

            console.log(`${player.name} played the word ${result.word} at position ${result.position}.`);
            console.log(`${player.name}'s words are now: ${player.words.join(', ')}`);
            console.log(`${player.name}'s letters are now: ${player.letters.join(', ')}`);

            // Calculate and display the score
            await calculateScore(player);
            await displayGameState();

        } else {
            console.log(`The word ${result.word} is not possible with the letters ${player.letters.join(', ')}.`);
/*             playAgain = false;
            await Promise.all([calculateScore(player), displayGameState()]); */
        }
    }
}
function jarnac(player) {
    // Ask for the line number
    let lineNumber = prompt("Enter the line number of the word you want to modify:");

    // Convert the line number to an integer
    lineNumber = parseInt(lineNumber, 10);

    // Validate the line number
    if (isNaN(lineNumber) || lineNumber < 1 || lineNumber > player.words.length) {
        console.error("Invalid line number");
        return;
    }

    // Ask for the new word
    let newWord = prompt("Enter the new word:");

    // Replace the word at the given line number with the new word
    player.words[lineNumber - 1] = newWord;
}

// Function to check if a word is possible with the given letters
function checkWord(letters, words, word, position) {
    // Create a copy of the letters array so we don't modify the original
    let wordArray = word.split("");
    let lettersCopy = [...letters];

    if (word.length < 3) {
        console.log(`The word ${word} is too short to be played.`);
        return false;
    }

    if (words.length > position) {
        let initWord = words[position];
        let initWordArray = initWord.split("");

        for (let letter of initWord) {
            let index = wordArray.indexOf(letter);
            if (index === -1) {
                // Letter not found in the array, or no more occurrences left, word is not possible
                console.log(`Letter ${letter} not found in lettersCopy. Word is not possible.`);
                return false;
            } else {
                // Remove only the first occurrence of the letter from the array
                wordArray.splice(wordArray.indexOf(letter), 1);
                initWordArray.splice(initWordArray.indexOf(letter), 1);
            }
        }
    } 
    
    let wordArrayFixed = [...wordArray]
     
    for (let letter of wordArrayFixed) {
        let index = letters.indexOf(letter);
        if (index === -1) {
            // Letter not found in the array, or no more occurrences left, word is not possible
            console.log(`Letter ${letter} not found in lettersCopy. Word is not possible.`);
            return false;
        } else {
            // Remove only the first occurrence of the letter from the array
            wordArray.splice(wordArray.indexOf(letter), 1);
            lettersCopy.splice(lettersCopy.indexOf(letter), 1);
        }
    }

    // All letters found, word is possible only if there are no remaining occurrences of letters
    return wordArray.length === 0;
}

function calculateScore(player) {
    return new Promise((resolve, reject) => {
        const points = [0, 0, 9, 16, 25, 36, 49, 64, 81];
        let score = 0;

        for (let word of player.words) {
            const length = word.length;
            score += points[length];
        }

        player.score = score;

    });
}


function calculateScore(player) {
    return new Promise((resolve, reject) => {
        // TODO: Calculate the score based on the player's words and letters
        // You can add your code here

        // For example, let's assume the score is the total number of letters in the player's words
        const score = player.words.reduce((total, word) => total + word.length, 0);

        // Update the player's score
        player.score = score;

        resolve();
    });
}

// Function to log a move to the game log file
function logMove(playerName, move) {
    const log = `${playerName} a joué le coup : ${move.word}\n`;
    return new Promise((resolve, reject) => {
        fs.appendFile(logFile, log, (err) => {
            if (err) reject(err);
            else resolve();
        });
    });
}

async function makeMove(playerName, move) {
    // Find the player in the game state
    const player = gameState.players.find((p) => p.name === playerName);

    // Update the player's move
    player.words[move.position] = move.word;

    for (let letter of move.word) {
        const index = player.letters.indexOf(letter);
        if (index !== -1) {
            player.letters.splice(index, 1);
        }
    }

    try {
        // Log the move
        await logMove(playerName, move);
    } catch (err) {
        console.error('Error logging move:', err);
    }
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
};


/* function addWord(playerName, word) {
    // Find the player
    const player = gameState.players.find(p => p.name === playerName);

    // Points for each word length
    const pointsArray = [0, 0, 9, 16, 25, 36, 49, 64, 81];

    // Calculate the points for the word
    const points = pointsArray[word.length];

    // Add the word to the player's words array
    player.words.push({ word, points });
}
 */
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
