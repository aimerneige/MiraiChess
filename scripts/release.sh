#!/usr/bin/bash
set -e

VERSION=$1
GREEN='\033[0;32m'
NC='\033[0m' # No Color
TARGET="MiraiChess-linux-amd64-${VERSION}"

printf "${GREEN}Start make release of version ${VERSION}.\n${NC}"

printf "${GREEN}Start building project.\n${NC}"
make all

printf "${GREEN}Start copy files.\n${NC}"
mkdir ./release/$TARGET

mkdir ./release/$TARGET/bin
cp ./bin/device ./release/$TARGET/bin
cp ./bin/inkscape ./release/$TARGET/bin
cp ./bin/mirai-chess-bot-linux-amd64-$VERSION ./release/$TARGET/bin

mkdir ./release/$TARGET/config
cp ./config/*.yaml ./release/$TARGET/config

mkdir ./release/$TARGET/logs

mkdir ./release/$TARGET/scripts
cp ./scripts/board2svg.py ./release/$TARGET/scripts
cp ./scripts/start.sh ./release/$TARGET

mkdir ./release/$TARGET/temp

cp LICENSE ./release/$TARGET/LICENSE

cp README.md ./release/$TARGET/README.md

echo "Here to get source code: https://github.com/aimerneige/MiraiChess." > ./release/$TARGET/src.txt

printf "${GREEN}Start compress release package.\n${NC}"
tar -C ./release/$TARGET -czvf ./release/$TARGET.tar.gz .

printf "${GREEN}Start clean temp files.\n${NC}"
rm ./release/$TARGET -rf

printf "${GREEN}Make release of version ${VERSION} successful.\n${NC}"
