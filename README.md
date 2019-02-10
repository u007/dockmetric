# dockmetric
docker telemetric saving

## usage

environment variables: 

* INFLUX_URL
* INFLUX_USER
* INFLUX_PASS

## build

env GOOS=linux GOARCH=amd64 go build -o dockmetric main.go
