set -e
-rf ./bin.inkscape
# get the latest release here: https://inkscape.org/release/
wget https://inkscape.org/gallery/item/33450/Inkscape-dc2aeda-x86_64.AppImage -O ./bin/inkscape
chmod +x ./bin/inkscape
