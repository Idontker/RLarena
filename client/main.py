

import sys
import tqdm
import requests
from strategy import *
import time


URL = "http://127.0.0.1:8081"
ROWS = 5
COLS = 3

TICKS = 500
GAMES = 100  # per matching

DEBUG = False


def moveToString(move):
    return "{}.{} - {}.{}".format(
        move["sourceRow"],
        move["sourceCol"],
        move["destRow"],
        move["destCol"],
    )


def movesToString(moves):
    return ", ".join(map(lambda move: "{}.{} - {}.{}".format(
        move["sourceRow"],
        move["sourceCol"],
        move["destRow"],
        move["destCol"],
    ), moves))


class Client():
    def __init__(self, username: str, strategy: 'Strategy'):
        self.strategy = strategy
        self.username = username
        self.token = ""

    def signUp(self):
        time.sleep(0.05)

        resp = requests.get(URL + "/user/signup?name={}".format(self.username))
        print("Sign up response:", resp.status_code, resp.text)
        if resp.status_code == 200:
            self.token = resp.json()["token"]
            return True
        print("Error during signUp:", resp.status_code, resp.text)
        return False

    def getActiveGames(self):
        resp = requests.get(URL + "/games/active/{}".format(self.token))
        if resp.status_code == 200:
            return resp.json()

        print("Error during 'GET /games/active/{}':".format(self.token),
              resp.status_code, resp.text)
        return None

    def getActiveGameIds(self):
        time.sleep(0.05)

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
        time.sleep(0.05)

        gameId = game["id"]
        game_state = game["game_state"]
        moveOptions = game_state["moveOptions"]

        selected = self.strategy.selectMove(moveOptions, game_state)
        return self.performAction(gameId, selected)

    def performAction(self, gameId, action):
        time.sleep(0.05)

        resp = requests.post(
            URL + "/game/{}/action?token={}".format(gameId, self.token), json=action)
        if resp.status_code == 200 or resp.status_code == 201:
            return True

        print("Error during 'POST /game/{}/action?token={}".format(gameId, self.token),
              resp.status_code, resp.text)
        return False

    def lookForGame(self):
        time.sleep(0.05)

        resp = requests.get(URL + "/match/queueup/{}".format(self.token))
        if resp.status_code == 200:
            return resp.text

        print("Error during 'GET /match/queueup/{}:".format(self.token),
              resp.status_code, resp.text)
        return None


if __name__ == "__main__":
    post = sys.argv[1] if sys.argv[1] else ""
    client1 = Client("radomMove" + post, RandomStrategy())
    client2 = Client("firstMove" + post, FirstMoveStrategy())
    client3 = Client("farwestMove" + post, FarwestStrategy())
    client4 = Client("alwaysTakeFirstMove" + post, AlwaysTakeFirstStrategy())
    client5 = Client("alwaysTakeRandomMove" + post, AlwaysTakeRandomStrategy())
    client6 = Client("AlwaysTakeFarwestStrategy" +
                     post, AlwaysTakeFarwestStrategy())

    clients = [client1, client2, client3, client4, client5, client6]
    for c in clients:
        c.signUp()

    for i in range(len(clients)):
        for j in range(i + 1, len(clients)):
            c1 = clients[i]
            c2 = clients[j]

            for _ in range(GAMES):
                out1 = c1.lookForGame()
                out2 = c2.lookForGame()
            print(GAMES, "games found for", c1.username, c2.username)

    ags = [client.getActiveGames() for client in clients]

    for client, ag in zip(clients, ags):
        print(client.username, "my trun :{} awating:{}".format(
            len(ag["my_turn"]), len(ag["awaiting"]))
        )

    tick = 0
    with tqdm.tqdm(total=TICKS) as pbar:
        while tick < TICKS and not all(len(ag["my_turn"]) == 0 and len(ag["awaiting"]) == 0 for ag in ags):
            for client, ag in zip(clients, ags):
                for game in ag["my_turn"]:
                    success = client.act(game)
                    if not success:
                        print()
                        print("Error during action for", client.username)
                        print(game)
                        print()

            ags = [client.getActiveGames() for client in clients]

            if DEBUG:
                for client, ag in zip(clients, ags):
                    print(tick, client.username, "my turn :{} awaiting:{}".format(
                        len(ag["my_turn"]), len(ag["awaiting"])
                    ))
                print("-" * 40)
            tick += 1
            pbar.update(1)
    # while tick < TICKS and not all(len(ag["my_turn"]) == 0 and len(ag["awaiting"]) == 0 for ag in ags):
    #     for client, ag in zip(clients, ags):
    #         for game in ag["my_turn"]:
    #             succes = client.act(game)
    #             if not succes:
    #                 print()
    #                 print("Error during action for", client.username)
    #                 print(game)
    #                 print()

    #     ags = [client.getActiveGames() for client in clients]
    #     for client, ag in zip(clients, ags):
    #         print(tick, client.username, "my turn :{} awaiting:{}".format(
    #             len(ag["my_turn"]), len(ag["awaiting"])
    #         ))
    #     print("-" * 40)
    #     tick += 1
    exit()

# resp = requests.get(URL + "/create?rows={}&cols={}".format(ROWS, COLS))
# print(resp.status_code)
# print(resp.text)

# id = resp.json()["id"]


# player1 = RandomStrategy()
# player2 = FirstMoveStrategy()

# finished = False
# i = 0
# while not finished:
#     resp_moves = requests.get(URL + "/moves?id={}".format(id))
#     resp_game_over = requests.get(URL + "/end?id={}".format(id))
#     resp_state = requests.get(URL + "/state?id={}".format(id))

#     # print("moves", resp_moves.status_code, resp_moves.text)
#     # print("state", resp_state.status_code, resp_state.text)
#     # print("game over", resp_game_over.status_code, resp_game_over.status_code)

#     state = resp_state.json()
#     moves = resp_moves.json()

#     if i % 2 == 0:
#         selected = player1.selectMove(moves, state)
#     else:
#         selected = player2.selectMove(moves, state)

#     displayGameState(state)
#     print("-" * 40)
#     print("game over?", resp_game_over.json())
#     print("-" * 40)
#     print("options:", movesToString(moves))

#     print("selected", moveToString(selected))
#     print("selected", selected)

#     action = requests.post(
#         URL + "/action?id={}".format(id), json=selected)
#     print("action accepted ?", action.status_code == 200,
#           "({} - {})".format(action.status_code, action.text))

#     # end turn
#     i += 1
#     resp_game_over = requests.get(URL + "/end?id={}".format(id))

#     finished = resp_game_over.json(
#     )["gameOver"] if resp_game_over.status_code == 200 else True


# resp_state = requests.get(URL + "/state?id={}".format(id))
# resp_game_over = requests.get(URL + "/end?id={}".format(id))

# displayGameState(resp_state.json())
# print("-" * 40)
# print("game over?", resp_game_over.json())
