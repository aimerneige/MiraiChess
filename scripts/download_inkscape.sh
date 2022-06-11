#!/usr/bin/bash
set -e

GREEN='\033[0;32m'
NC='\033[0m' # No Color
INKSCAPE_BIN='./bin/inkscape'
INKSCAPE_DOWNLOAD_LINK='https://inkscape.org/gallery/item/33450/Inkscape-dc2aeda-x86_64.AppImage' # get the latest release here: https://inkscape.org/release/

if [ -f $INKSCAPE_BIN ]
then
    printf "${GREEN}Inkscape alreade exist. Download skiped.\n${NC}"
else
    printf "${GREEN}Start download Inkscape......\n${NC}"
    wget $INKSCAPE_DOWNLOAD_LINK -O $INKSCAPE_BIN
    chmod +x $INKSCAPE_BIN
fi
