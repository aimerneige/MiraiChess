#!/usr/bin/bash
set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color
DEVICE_FILE="./device.json"

./scripts/download_inkscape.sh

if [ -f $DEVICE_FILE ]
then
    printf "${GREEN}Device file already exist. Skiped.\n${NC}"
else
    printf "${GREEN}Start to generate devive file.\n${NC}"
    ./bin/device
fi

./bin/bot
