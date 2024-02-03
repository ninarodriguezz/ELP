const prompt = require('prompt');
const util = require('util');
const fs = require('fs');
const logFile = 'game_log.txt';

prompt.start();
const get = util.promisify(prompt.get);

function checkWord(letters, words, word, position) {
    let wordArray = word.split("");
    let lettersCopy = [...letters];

    if (word.length < 3) {
        console.log(`The word ${word} is too short to be played.`);
        return [false, []];
    }

    if (words.length >= position) {
        var initWord = words[position - 1];
        var initWordArray = initWord.split("");

        for (let letter of initWord) {
            var index = wordArray.indexOf(letter);
            if (index === -1) {
                console.log(`Letter ${letter} not found in ${word}. Word is not possible.`);
                return [false, []];
            } else {
                wordArray.splice(wordArray.indexOf(letter), 1);
                initWordArray.splice(initWordArray.indexOf(letter), 1);
            }
        }

        if (initWordArray.length != 0) {
            console.log(`To write the word ${word}, you are not using all the letters of the word ${initWord}.`);
            return [false, []];
        }
    } 
    
    var wordArrayFixed = [...wordArray];
    var lettersUsed = [];
     
    for (let letter of wordArrayFixed) {
        var index = letters.indexOf(letter);
        if (index === -1) {
            console.log(`Letter ${letter} not found in your letters. Word is not possible.`);
            return [false, []];
        } else {
            lettersUsed.push(letter);
            wordArray.splice(wordArray.indexOf(letter), 1);
            lettersCopy.splice(lettersCopy.indexOf(letter), 1);
        }
    }
    console.log(`The word ${word} is possible with the letters ${letters.join(', ')}.`);
    return [wordArray.length === 0, lettersUsed];
}

async function makeMove(player, move, lettersUsed) {
    player.words[move.position - 1] = move.word;

    for (let letter of lettersUsed) {
        const index = player.letters.indexOf(letter);
        if (index !== -1) {
            player.letters.splice(index, 1);
        }
    }

    try {
        await logMove(player, move);
    } catch (err) {
        console.error('Error logging move:', err);
    }
}

function logMove(player, move) {
    const log = `${player.name} played the word : ${move.word} in line ${move.position}\n`;
    return new Promise((resolve, reject) => {
        fs.appendFile(logFile, log, (err) => {
            if (err) reject(err);
            else resolve();
        });
    });
}

/* ...................JARNAC............................... */

async function jarnac(player, otherPlayer) {
    const lineNumberResult = await get([{
        name: 'lineNumber',
        description: 'Enter the line number of the word you want to modify',
        type: 'number',
        required: true
    }]);

    let lineNumber = lineNumberResult.lineNumber;

    while (isNaN(lineNumber) || lineNumber < 1 || (lineNumber > otherPlayer.words.length && otherPlayer.words.length !== 0)) {
        const lineNumberResult = await get([{
            name: 'lineNumber',
            description: 'Enter the line number of the word you want to modify',
            type: 'number',
            required: true
        }]);

        lineNumber = lineNumberResult.lineNumber;

        if (isNaN(lineNumber) || lineNumber < 1 || (lineNumber > otherPlayer.words.length && otherPlayer.words.length !== 0)) {
            console.error("Invalid line number");
        }
    }

    const newWordResult = await get([{
        name: 'newWord',
        description: 'Enter the new word',
        type: 'string',
        required: true
    }]);

    let newWord = newWordResult.newWord.toUpperCase();
    if (otherPlayer.words[lineNumber - 1]) {
        if (newWord.length <= otherPlayer.words[lineNumber - 1].length) {
            console.error("The new word must be longer than the existing word");
            return;
    }
    }

    var [isPossible, lettersUsed] = checkWord(otherPlayer.letters, otherPlayer.words, newWord, lineNumber);
    if (isPossible) {
        const move = { word: newWord, position: lineNumber};
        await makeJarnacMove(player, otherPlayer, move, lettersUsed);

        console.log(`${player.name} modified the word at line ${lineNumber} to ${newWord}.`);
        console.log(`${otherPlayer.name}'s words are now: ${otherPlayer.words.join(', ')}`);
    } else {
        console.log(`The word ${newWord} is not possible with the letters ${otherPlayer.letters.join(', ')}.`);
    }
}

async function makeJarnacMove(player, otherPlayer, move, lettersUsed) {
    var word = move.word;
    const wordIndex = move.position - 1;
    if (otherPlayer.words){ 
        otherPlayer.words.splice(wordIndex, 1);
    }
    player.words.push(word);

    for (let letter of lettersUsed) {
        const index = otherPlayer.letters.indexOf(letter);
        if (index !== -1) {
            otherPlayer.letters.splice(index, 1);
        }
    }

    console.log(`${player.name} has taken the word ${word} from ${otherPlayer.name}.`);
    console.log(`${player.name}'s words are now: ${player.words.join(', ')}`);
    console.log(`${otherPlayer.name}'s words are now: ${otherPlayer.words.join(', ')}`);
    console.log(`${otherPlayer.name}'s letters are now: ${otherPlayer.letters.join(', ')}`);

    try {
        await logJarnac(player, otherPlayer, move);
    } catch (err) {
        console.error('Error logging move:', err);
    }
}

function logJarnac(player, otherPlayer, move) {
    const log = `JARNAC! ${player.name} has stolen the word ${move.word} from ${otherPlayer.name} in line ${move.position}\n`;
    return new Promise((resolve, reject) => {
        fs.appendFile(logFile, log, (err) => {
            if (err) reject(err);
            else resolve();
        });
    });
}
module.exports = {
    checkWord,
    makeMove,
    logMove,
    jarnac,
    makeJarnacMove
};