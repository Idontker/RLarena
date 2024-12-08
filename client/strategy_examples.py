from abc import ABC
import random
from strategy import Strategy


class RandomStrategy(Strategy):
    def selectMove(self, moves, state):
        return random.choice(moves)


class FirstMoveStrategy(Strategy):
    def selectMove(self, moves, state):
        return moves[0]


class FarwestStrategy(Strategy):
    def selectMove(self, moves, state):
        player = moves[0]["player"]

        increasing = player == 1

        farwest = 0 if increasing else state["rows"] - 1
        options = []

        for move in moves:
            if increasing:
                if move["destRow"] > farwest:
                    farwest = move["destRow"]
                    options = [move]
                elif move["destRow"] == farwest:
                    options.append(move)
            else:
                if move["destRow"] < farwest:
                    farwest = move["destRow"]
                    options = [move]
                elif move["destRow"] == farwest:
                    options.append(move)

        return random.choice(options)
        # TODO: Try always to take (if multiple, random amongst them)


class AlwaysTakeFirstStrategy(Strategy):
    def selectMove(self, moves, state):
        options = []

        for move in moves:
            if move["destCol"] != move["sourceCol"]:
                options.append(move)

        if len(options) == 0:
            return moves[0]
        return options[0]


class AlwaysTakeRandomStrategy(Strategy):
    def selectMove(self, moves, state):

        options = []

        for move in moves:
            if move["destCol"] != move["sourceCol"]:
                options.append(move)

        if len(options) == 0:
            return random.choice(moves)
        return random.choice(options)


class AlwaysTakeFarwestStrategy(Strategy):
    def selectMove(self, moves, state):
        player = moves[0]["player"]
        increasing = player == 1

        takeOptions = []

        for move in moves:
            if move["destCol"] != move["sourceCol"]:
                takeOptions.append(move)

        farwest = 0 if increasing else state["rows"] - 1

        options = []
        for move in takeOptions:
            if increasing:
                if move["destRow"] > farwest:
                    farwest = move["destRow"]
                    options = [move]
                elif move["destRow"] == farwest:
                    options.append(move)
            else:
                if move["destRow"] < farwest:
                    farwest = move["destRow"]
                    options = [move]
                elif move["destRow"] == farwest:
                    options.append(move)

        if len(options) == 0:
            farwest = 0 if increasing else state["rows"] - 1

            options = []
            for move in moves:
                if increasing:
                    if move["destRow"] > farwest:
                        farwest = move["destRow"]
                        options = [move]
                    elif move["destRow"] == farwest:
                        options.append(move)
                else:
                    if move["destRow"] < farwest:
                        farwest = move["destRow"]
                        options = [move]
                    elif move["destRow"] == farwest:
                        options.append(move)

        return random.choice(options)

    # TODO: Try never to be taken (if multiple, random amongst them)
