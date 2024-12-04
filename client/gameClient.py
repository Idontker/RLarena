import requests
import time
from strategy import Strategy

URL = "http://127.0.0.1:8081"
SLEEP_TIME = 0.01


class Client():
    def __init__(self, username: str, strategy: Strategy):
        self.strategy = strategy
        self.username = username
        self.token = ""

    def signUp(self):
        time.sleep(SLEEP_TIME)

        resp = requests.get(URL + "/user/signup?name={}".format(self.username))
        print("Sign up response:", resp.status_code, resp.text)
        if resp.status_code == 200:
            self.token = resp.json()["token"]
            return True
        print("Error during signUp:", resp.status_code, resp.text)
        return False

    def getActiveGames(self):
        time.sleep(SLEEP_TIME)

        resp = requests.get(URL + "/games/active/{}".format(self.token))
        if resp.status_code == 200:
            return resp.json()

        print("Error during 'GET /games/active/{}':".format(self.token),
              resp.status_code, resp.text)
        return None

    def getActiveGameIds(self):
        ag = self.getActiveGames()
        if ag is None:
            return None

        return {
            "my_turn": [game["id"] for game in ag["my_turn"]],
            "awaiting": [game["id"] for game in ag["awaiting"]]
        }

    def printGameState(self, gameId):
        resp = requests.get(URL + "/game/{}/state".format(gameId))
        if resp.status_code == 200:
            game = resp.json()
            print("({})Game state: {} vs {}".format())

            game_state = game["game_state"]
            for row in reversed(game_state["board"]):
                print(row)
            print("History:")
            for turn in game_state["history"]:
                print("Player {} moved ({}.{}) to ({}.{})".format(
                    turn["player"],
                    turn["sourceRow"],
                    turn["sourceCol"],
                    turn["destRow"],
                    turn["destCol"]
                ))
            return True

        print("Error during 'GET /game/{}/state:".format(gameId),
              resp.status_code, resp.text)
        return False

    def getGame(self, gameId):
        resp = requests.get(URL + "/game/{}/state".format(gameId))
        if resp != 200:
            return resp.json()

        print("Error during 'GET /game/{}/state:".format(gameId),
              resp.status_code, resp.text)
        return None

    def act(self, game):
        gameId = game["id"]
        game_state = game["game_state"]
        moveOptions = game_state["moveOptions"]

        selected = self.strategy.selectMove(moveOptions, game_state)
        return self.performAction(gameId, selected)

    def actBulk(self, games):
        time.sleep(SLEEP_TIME)
        payloads = []

        for game in games:
            gameId = game["id"]
            game_state = game["game_state"]
            moveOptions = game_state["moveOptions"]

            selected = self.strategy.selectMove(moveOptions, game_state)
            payloads.append({
                "gameId": gameId,
                "action": selected
            })

        resp = requests.post(
            URL + "/games/actions?token={}".format(self.token), json=payloads)
        if resp.status_code == 200 or resp.status_code == 201:
            return True

        print("Error during 'POST /games/actions?token={}".format(self.token),
              resp.status_code, resp.text)
        return False

    def performAction(self, gameId, action):
        time.sleep(0.05)

        resp = requests.post(
            URL + "/game/{}/action?token={}".format(gameId, self.token), json=action)
        if resp.status_code == 200 or resp.status_code == 201:
            return True

        print("Error during 'POST /game/{}/action?token={}".format(gameId, self.token),
              resp.status_code, resp.text)
        return False

    def lookForGame(self,):
        time.sleep(0.05)

        resp = requests.get(URL + "/match/queueup/{}".format(self.token))
        if resp.status_code == 200:
            return resp.text

        print("Error during 'GET /match/queueup/{}:".format(self.token),
              resp.status_code, resp.text)
        return None
