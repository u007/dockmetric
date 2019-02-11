#!/bin/bash

# [ -f /custom/installed ] && echo "already setup" || bash /setup.sh &
bash /setup.sh &

# echo "Starting: $CONTAINERS..."
# touch $HOME/nohup.out
# nohup dockmetric $CONTAINERS & tail -f $HOME/nohup.out &
echo "Starting grafana..."
bash /run.sh
