#!/bin/bash

curl -s $INFLUX_URL > /dev/null
while [ $? -ne 0 ]; do
  echo "waiting..."
  sleep 1
  curl -s $INFLUX_URL > /dev/null
done

echo "Starting metrics: $CONTAINERS..."
dockmetric $CONTAINERS
