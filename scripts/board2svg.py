#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import sys
import chess
import chess.svg


def generate_board_svg(fen_str: str, file_path: str, last_move_uci: str):
    board = chess.Board(fen_str)
    if board.is_check():
        king_square_index = board.king(board.turn)
        king_square_name = chess.square_name(king_square_index)
        king_square = chess.parse_square(king_square_name)
    else:
        king_square = None
    last_move = chess.Move.from_uci(last_move_uci)
    arrows = [chess.svg.Arrow(last_move.from_square, last_move.to_square, color="#7D6C46BB")]
    svg = chess.svg.board(
        board=board,
        orientation=board.turn,
        lastmove=last_move,
        check=king_square,
        arrows=arrows,
        size=720,
        coordinates=True,
    )
    with open(file_path, 'w') as f:
        f.write(svg)


if __name__ == "__main__":
    args = sys.argv
    if len(args) != 4:
        print("wrong args")
        exit(1)
    generate_board_svg(fen_str=args[1], file_path=args[2], last_move_uci=args[3])
    exit(0)
