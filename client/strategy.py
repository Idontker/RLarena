from abc import ABC
import random
from dataclasses import dataclass
from typing import List, Dict, Optional

random.seed = 42


class Strategy(ABC):
    def selectMove(self, moves: 'List[Move]', state: 'GameState') -> 'Move':
        pass


@dataclass
class Move:
    turnID: int
    destRow: int
    destCol: int
    sourceRow: int
    sourceCol: int
    player: int


@dataclass
class GameState:
    rows: int
    cols: int
    history: List[Move]

    # 2d array. 0 is empty, 1 is player 1, 2 is player 2
    board: List[List[int]]
    gameOver: bool

    # 0 is no winner, 1 is player 1, 2 is player 2
    # -1 draw
    winner: int
    moveOptions: List[Move]

    # 1 or 2, depending on which player's turn it is
    currentPlayer: int
