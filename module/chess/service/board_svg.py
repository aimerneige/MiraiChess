#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import sys
import chess
import chess.svg


def generate_board_svg(fen_str: str, file_path: str):
    board = chess.Board(fen_str)
    svg = chess.svg.board(
        board=board,
        size=720,
        orientation=board.turn,
    )
    with open(file_path, 'w') as f:
        f.write(svg)


if __name__ == "__main__":
    args = sys.argv
    if len(args) != 3:
        print("wrong args")
        exit(1)
    generate_board_svg(fen_str=args[1], file_path=args[2])
    exit(0)
