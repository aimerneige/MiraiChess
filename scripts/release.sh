#!/usr/bin/bash
set -e

VERSION=$1
GREEN='\033[0;32m'
NC='\033[0m' # No Color
TARGET="MiraiChess-linux-amd64-${VERSION}"

printf "${GREEN}Start make release of version ${VERSION}.\n${NC}"
mkdir ./release/$TARGET

mkdir ./release/$TARGET/bin
cp ./bin/device ./release/$TARGET/bin
cp ./bin/mirai-chess-bot-linux-amd64-$VERSION ./release/$TARGET/bin/bot

mkdir ./release/$TARGET/config
cp ./config/*.yaml ./release/$TARGET/config

mkdir ./release/$TARGET/db

mkdir ./release/$TARGET/logs

mkdir ./release/$TARGET/scripts
cp ./scripts/download_inkscape.sh ./release/$TARGET/scripts
cp ./scripts/start.sh ./release/$TARGET
cp ./scripts/update.sh ./release/$TARGET

mkdir ./release/$TARGET/temp

cp LICENSE ./release/$TARGET/LICENSE

cp mirai-chess.service ./release/$TARGET/mirai-chess.service

cp README.md ./release/$TARGET/README.md

echo "Here to get source code: https://github.com/aimerneige/MiraiChess." > ./release/$TARGET/src.txt

tar -C ./release/$TARGET -czvf ./release/$TARGET.tar.gz .

printf "${GREEN}Start clean temp files.\n${NC}"
rm ./release/$TARGET -rf

printf "${GREEN}Make release of version ${VERSION} successful.\n${NC}"
