# Jarnac Game Readme

## Introduction

This repository contains a Jarnac word game implemented in Node.js. The game allows two players to take turns forming a maximum of 8 words from a set of letters. The first player to write 8 words make the game stop. The winner is the player with the most points cumulated.

## Files

This project contains the 3 following files :
    
    - game.js: This file contains the main game logic, including the game loop and the functions to start and set up the game.
    - move.js: This file contains the logic for making moves in the game, including checking if a word can be played, playing a word and performing a "jarnac".
    - state.js: This file contains the shared data gameState and the function which print the board.

## Getting Started

### Prerequisites

Before running the game, make sure you have Node.js installed on your machine.

### Installation

1. Clone this repository to your local machine.
2. Navigate to the repository's root directory.
3. Run the command `npm install` to install the required dependencies.

## Usage

To start the game, run the main script named game.js:

```bash
node game.js
```

All interaction with the game is done through the terminal. So you can then enter the players' names in the terminal and start to play to Jarnac. 
Both players will take turns using the same keyboard.

Notice that all the moves made during a game will be written in the file 'game_log.txt'.
