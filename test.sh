#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd $SCRIPT_DIR

cd cmd/gox || exit
./install.sh
cd ../../

# cd tutorial/06-Demos || exit
cd tutorial/05-Animation || exit

pkill -f gdspx_web_server.py
spx clear
spx runweb -serveraddr "0:$1"
cd ../../