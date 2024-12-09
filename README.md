# RL Arena
![image | width=30%](https://github.com/user-attachments/assets/44f12e8e-049d-4ec4-b0f6-74b4ed170ec7)

This project contains a pawn-chess (dt. "Bauernschach") server for learning RL (reinforcement learning) and competing against other and their bots.
It is live 🥳🥳🚀
https://rlarena.akbakas.de/game?id=2
## How to play (Step by Step Guide)
Hi there, this is a detailed step-by-step guide on how to create your own pawn-chess bot and play. If the description is too easy / slow, feel free to only read the code segments. They might be self explanatory.
### Requirments 📚
- Python 3.X
- git (optional)
### Setup 🚀
#### Download Repository
1. Option: (recommended) Download this repo via git or
2. Option: Download from [here](https://github.com/Idontker/RLarena)  by clicking "Code" and then "Download Zip". Extract the repo locally.
![image](https://github.com/user-attachments/assets/40534e8c-6f50-4ee6-993a-845de31b3ad5)

#### Install Dependencies
Sadly, there is one external dependency for the python client ([requests](https://pypi.org/project/requests/)). It is a very popular package that implements HTTP Requests.
To install `requests` for the user (recommended) perform:
```bash
pip install --user requests
```
If you want to create a virtual python envoirment instead, use your favorite venv package (e.g. venv, pipenv, conda, pdm).

### Start your first Bot 🤖
Navigate to the `./client` folder
```bash
cd client
python p1.py
```
If you are not running your server locally, this will lead to an error. Open the p1.py file and modify the commented lines to match: 
```python
# p1.py

URLBASE = "https://rlarena.akbakas.de/"
client = Client("radomMove", RandomStrategy(),urlbase=URLBASE)

# client = Client("radomMove", RandomStrategy())
```
Now, starting  your bot with `python p1 .py` will work fine. But, if noone else has a bot currently running, your bot will get bored quickly by running out of turns to take.
Therefore, consider running a second bot in parallel (by e.g. opening another shell and exceuting
```
python p2.py
```

### Developing your own Bot 🦆🤖🦆
The gameClient is basicly finished and the "main" is done as well (`p1.py`). The only thing missing is your strategy: How shall your bot chose the next move to perform.
Create a new file `myStrat.py` and paste in the following code:
```python
# myStrat.py

import random
from strategy import Strategy, Move, GameState
from typing import List


class MyStrat(Strategy):
    def selectMove(self, moves: List[Move], state: GameState) -> Move:
        # TODO:
        pass
```
This is (with one minor exception) the only file, in which you need to write code. You are given a list of valid moves and the current gamestate and shall return the move, you want your bot to play.
A simple strategy might be, to pick random
```python
def selectMove(self, moves: List[Move], state: GameState) -> Move:
    my_pick = random.choice(moves)
    return
```
or always pick the first move:
```python
def selectMove(self, moves: List[Move], state: GameState) -> Move:
    my_pick = moves[0]
    return
```
The `Move` and `GameState` class/interface is defined within Strategy

See [`strategy_examples.py`](https://github.com/Idontker/RLarena/blob/main/client/strategy_examples.py) contains more basic examples for different bots. 
Currently, none of them are using RL - that is your job ;)

### Runing your bot 🚀🤖
Copy `p1.py` or simply modify it, to import and use your strategy instead of the `RandomStrategy`: 
```python
# p1.py
from myStrat import MyStrat
URLBASE = "https://rlarena.akbakas.de/"
client = Client("myLovleyBot", MyStrat(),urlbase=URLBASE)

# DO NOT FORGET TO CHANGE YOUR BOTS NAME! OTHERWISE, IT WILL
# PLAY ON THE SAME ACOUNT 
# ... (rest of the code) 
```



## Game Server
### Database
Create migration:
```
goose -dir=migrations create rlarena sql
```
Create Database from migrations
```
goose -dir=migrations sqlite3 app.db up
```
Delete Database from migrations
```
goose -dir=migrations sqlite3 app.db down
```
