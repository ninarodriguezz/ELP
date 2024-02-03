let gameState = {
    players: [],
    currentPlayer: 0,
};

getGameState = () => {
    return gameState;
};  

setGameState = (newState) => {
    gameState = newState;   
};  

displayGameState = () => {
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



module.exports = {
    getGameState: () => gameState,
    setGameState: (newState) => { gameState = newState; },
    displayGameState,
};