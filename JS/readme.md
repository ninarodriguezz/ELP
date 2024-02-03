# Jarnac Game Readme

## Introduction

This repository contains a Jarnac word game implemented in Node.js. The game allows two players to take turns forming a maximum of 8 words from a set of letters. The first player to write 8 words make the game stop. The winner is the player with the most points cumulated.

## Files

This project contains the 5 following files :
    
    - game.js: This file contains the main game logic, including the game loop and the functions to start and set up the game.
    - move.js: This file contains the logic for making moves in the game, including playing a word, passing a turn, and performing a "jarnac".
    - player.js: This file defines the Player class, which represents a player in the game. Each player has a name, a set of letters, and a set of words they have played.
    - letter.js: This file contains the logic for drawing letters and managing the letter pool.
    - word.js: This file contains the logic for managing words, including checking if a word is valid and adding a word to a player's set of words.

## Getting Started

### Prerequisites

Before running the game, make sure you have Node.js installed on your machine.

### Installation

1. Clone this repository to your local machine.
2. Navigate to the repository's root directory.
3. Run the command `npm install` to install the required dependencies.

## Usage

To start the game, run the main script:

```bash
node main.js
```

All interaction with the game is done through the terminal. So you can then enter the players' names in the terminal and start to play to Jarnac.