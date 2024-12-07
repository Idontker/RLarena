

import sys
import tqdm
from strategy import *
from gameClient import Client

TICKS = 20
GAMES = 5  # per matching

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


if __name__ == "__main__":
    post = sys.argv[1] if sys.argv[1] else ""
    client1 = Client("radomMove" + post, RandomStrategy())
    client2 = Client("firstMove" + post, FirstMoveStrategy())

    clients = [client1, client2]
    for c in clients:
        c.signUp()

    for i in range(len(clients)):
        for j in range(i + 1, len(clients)):
            c1 = clients[i]
            c2 = clients[j]

            # for _ in range(GAMES):
            out1 = c1.lookForGame(gameCount=GAMES)
            out2 = c2.lookForGame(gameCount=GAMES)
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
                games = ag["my_turn"]
                success = client.actBulk(games)

            ags = [client.getActiveGames() for client in clients]
            tick += 1
            pbar.update(1)

            if DEBUG:
                for client, ag in zip(clients, ags):
                    print(tick, client.username, "my turn :{} awaiting:{}".format(
                        len(ag["my_turn"]), len(ag["awaiting"])
                    ))
                print("-" * 40)

    exit()
