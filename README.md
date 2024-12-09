# RL Arena
This project contains a pawn-chess (dt. "Bauernschach") server for learning RL (reinforcement learning) and competing against other and their bots.
## How to play
### Requirments
- Python 3.X
- git (optional)
### Setup 
#### Download Repository
1. Option: (recommended) Download this repo via git or
2. Option: Download from [here](https://github.com/Idontker/RLarena)  by clicking "Code" and then "Download Zip". Extract the repo locally.
![image](https://github.com/user-attachments/assets/40534e8c-6f50-4ee6-993a-845de31b3ad5)

#### Install Dependencies
Sadly, there is one external dependency for the python client (requests). It is a very popular package that implements HTTP Requests.
To install `requests` for the user (recommended) perform:
```bash
pip install --user requests
```
If you want to create a virtual python envoirment instead, use your favorite venv package (e.g. venv, pipenv, conda, pdm).

### Start your first Bot
Navigate to the `./client` folder
```bash
cd client
python p1.py
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
