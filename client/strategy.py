from abc import ABC
import random

random.seed = 42


class Strategy(ABC):
    def selectMove(self, moves, state):
        pass
