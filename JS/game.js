const prompt = require('prompt');
const util = require('util');
const fs = require('fs');
const logFile = 'game_log.txt';
const mv = require('./move');
const { getGameState, setGameState, displayGameState } = require('./state');

prompt.start();
const get = util.promisify(prompt.get);

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
            words: []
        };
        var gameState = getGameState();
        gameState.players.push(player);
        console.log(`Player ${result.playerName} has joined the game`);
        if (gameState.players.length < 2) {
            getPlayerName(2);
        } else {

            setupGame(gameState.players.map(p => p.name));
        }
        setGameState(gameState);
    });
}

const letters = 'A'.repeat(14) + 'B'.repeat(4) + 'C'.repeat(7) + 'D'.repeat(5) + 'E'.repeat(19) + 'F'.repeat(2) + 'G'.repeat(4) + 'H'.repeat(2) + 'I'.repeat(11) + 'J'.repeat(1) + 'K'.repeat(1) + 'L'.repeat(6) + 'M'.repeat(5) + 'N'.repeat(9) + 'O'.repeat(8) + 'P'.repeat(4) + 'Q'.repeat(1) + 'R'.repeat(10) + 'S'.repeat(7) + 'T'.repeat(9) + 'U'.repeat(8) + 'V'.repeat(2) + 'W'.repeat(1) + 'X'.repeat(1) + 'Y'.repeat(1) + 'Z'.repeat(2);
const lettersArray = letters.split('');


function drawLetter() {
    const index = Math.floor(Math.random() * lettersArray.length);
    return lettersArray.splice(index, 1)[0];
}

/**
 * Sets up the game with the given player names.
 * 
 * This function initializes the game state, creates a player object for each player name, 
 * draws 5 letters for each player, and starts the game.
 * 
 * @param {Array} playerNames - An array of strings, each representing a player's name.
 * 
 * @returns {void}
 */

function setupGame(playerNames) {
    let gameState = getGameState();
    for (const playerName of playerNames) {
        let player = gameState.players.find(p => p.name === playerName);

        for (let i = 0; i < 5; i++) {
            player.letters.push(drawLetter());
        }
    }
    setGameState(gameState);
    startGame(gameState);

}

/**
 * Starts the game with the given game state.
 * 
 * This function initializes the game, sets the current player, and enters a loop that continues until the game ends. 
 * In each iteration of the loop, it displays the game state, sets a 60-second timer for the player's turn, 
 * and handles the player's turn. If the player does not complete their turn within 60 seconds, 
 * the function catches the timeout error and moves on to the next player. 
 * After each turn, it checks if the current player wants to perform a "jarnac" and handles it if so. 
 * It also logs the end of each turn to a file.
 * 
 * @param {Object} gameState - The state of the game. This object should have a 'players' property that is an array of player objects, 
 * and a 'currentPlayer' property that is the index of the current player in the 'players' array.
 * 
 * @returns {Promise} - A promise that resolves when the game ends.
 * 
 * @async
 */
async function startGame(gameState) {
    gameState.currentPlayer = 0;
    let running = true;
    if (!gameState.players || !Array.isArray(gameState.players)) {
        console.error('gameState.players is not an array');
        return;
    }
    while (running) {
        displayGameState();
        console.log("You have 60 seconds to complete your turn."); // Added line
        const timer = new Promise((resolve, reject) => {
            setTimeout(() => {
                reject(new Error('Time out'));
            }, 60000); 
        });
        try {
            await Promise.race([playerTurn(gameState.players[gameState.currentPlayer]), timer]);
        } catch (error) {
            console.error(error.message);
            gameState.currentPlayer = (gameState.currentPlayer + 1) % gameState.players.length;
            setGameState(gameState);
            continue; 
        }
        gameState.currentPlayer = (gameState.currentPlayer + 1) % gameState.players.length;
        setGameState(gameState);
        const jarnacResult = await get([{
            name: 'jarnac',
            description: gameState.players[gameState.currentPlayer].name + ', do you want to do a "jarnac"? (yes/no)',
            type: 'string',
            required: true
        }]);
        if (jarnacResult.jarnac.toLowerCase() === 'yes') {
            var otherPlayer = (gameState.currentPlayer + 1) % gameState.players.length;
            await mv.jarnac(gameState.players[gameState.currentPlayer], gameState.players[otherPlayer])
        }
        fs.appendFileSync('game_log.txt', '----------------------\n');
    }
}

/**
 * Checks if the end condition of the game has been met.
 * 
 * This function retrieves the current game state and checks if any player has played 8 words. 
 * If a player has 8 words, the game ends and the function returns true. 
 * If no player has 8 words, the function returns false.
 * 
 * @returns {boolean} - Returns true if the end condition of the game has been met, false otherwise.
 */
function checkEndCondition() {
    // ... function body ...
}
function checkEndCondition() {
    gameState = getGameState()
    for (let player of gameState.players) {
        if (player.words.length === 8) {
            return true;
        }
    }
    return false;
}

function determineWinner() {
    gameState = getGameState();
    let highestScore = 0;
    let winner = null;

    for (let player of gameState.players) {
        let score = player.score; 
        if (score > highestScore) {
            highestScore = score;
            winner = player;
        }
    }

    return winner;
}

/**
 * Handles a player's turn in the game.
 * 
 * This function manages the actions a player can take during their turn, including drawing a letter, choosing to play a word or pass their turn, 
 * validating the chosen action, and playing a word if the action is 'play'. It also checks if the game has ended after each turn.
 * 
 * @param {Object} player - The player who is currently taking their turn. The player object should have properties for 'letters' (an array of letters the player can play), 
 * 'words' (an array of words the player has played), and 'name' (the player's name).
 * 
 * @returns {Promise} - A promise that resolves when the player's turn is over. If the game ends during the player's turn, the function will return immediately.
 * 
 * @async
 */
async function playerTurn(player) {
    let playAgain = true;

    while (playAgain) {
        let validAction = false;
        player.letters.push(drawLetter());

        while (!validAction) {
            const actionResult = await get([{
                name: 'action',
                description: `${player.name}, do you want to play a word or pass your turn? (play/pass)`,
                type: 'string',
                required: true
            }]);

            if (actionResult.action.toLowerCase() === 'pass') {
                console.log(`${player.name} has decided to pass his turn.`);
                playAgain = false;
                validAction = true;
            } else if (actionResult.action.toLowerCase() === 'play') {
                validAction = true;
            } else {
                console.log('Invalid action. Please enter "play" or "pass".');
            }
        }

        if (!playAgain) {
            break;
        }

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
                var maxPosition = player.words.length + 1;
                return value >= 1 && value <= maxPosition;
            },
            message: 'Position must be between 1 and ' + (player.words.length + 1)
        }]);
        const word = result.word.toUpperCase();
        var [isPossible, lettersUsed] = mv.checkWord(player.letters, player.words, word, result.position);
        if (isPossible) {
            const move = { word: word, position: result.position };
            await mv.makeMove(player, move, lettersUsed);

            console.log(`${player.name} played the word ${result.word} at position ${result.position}.`);
            console.log(`${player.name}'s words are now: ${player.words.join(', ')}`);
            console.log(`${player.name}'s letters are now: ${player.letters.join(', ')}`);

            await calculateScore(player);
            await displayGameState();

            if (checkEndCondition()) {
                let winner = determineWinner();
                console.log(`The game has ended. The winner is ${winner.name}.`);
                return;
            }
        } else {
            console.log(`The word ${result.word} is not possible with the letters ${player.letters.join(', ')}.`);
        }
    }

}

function calculateScore(player) {
    return new Promise((resolve, reject) => {
        try {
            const points = [0, 0, 0, 9, 16, 25, 36, 49, 64, 81];
            let score = 0;

            for (let word of player.words) {
                const length = word.length;
                score += points[length];
            }

            player.score = score;
            resolve(); 
        } catch (error) {
            reject(error); 
        }
    });
}

cleanFile();
getPlayerName(1);

module.exports = {
    onErr,
    cleanFile,
    getPlayerName, 
    setupGame,
    startGame,
};