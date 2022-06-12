#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import sys
import chess
import chess.svg


def generate_board_svg(fen_str: str, file_path: str, last_move_uci: str):
    board = chess.Board(fen_str)

    king_square = None
    if board.is_check():
        king_square_index = board.king(board.turn)
        king_square_name = chess.square_name(king_square_index)
        king_square = chess.parse_square(king_square_name)

    last_move = None
    if last_move_uci != "None":
        last_move = chess.Move.from_uci(last_move_uci)

    svg = chess.svg.board(
        board=board,
        orientation=board.turn,
        lastmove=last_move,
        check=king_square,
        size=720,
        coordinates=True,
    )
    with open(file_path, 'w') as f:
        f.write(svg)


if __name__ == "__main__":
    args = sys.argv
    if len(args) != 4:
        print("Wrong args")
        exit(1)
    fen_str = args[1]
    file_path = args[2]
    last_move_uci = args[3]
    generate_board_svg(fen_str, file_path, last_move_uci)
    exit(0)
