#!/usr/bin/bash
set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color
DEVICE_FILE="./device.json"

DIR="/opt/MiraiChess"

printf "${GREEN}Stop mirai-chess.service.\n${NC}"
systemctl stop mirai-chess.service

printf "${GREEN}Start to update ./bin/bot.\n${NC}"
rm $DIR/bin/bot
cp ./bin/bot $DIR/bin/bot

printf "${GREEN}Start to update README.md.\n${NC}"
rm $DIR/README.md
cp ./README.md $DIR/README.md

printf "${GREEN}Start mirai-chess.service.\n${NC}"
systemctl start mirai-chess.service

printf "${GREEN}Update Finish.\n${NC}"
