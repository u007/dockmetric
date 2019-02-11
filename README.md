# dockmetric

docker telemetric saving

![containers](https://raw.githubusercontent.com/u007/dockmetric/master/images/containers.png)
![detail](https://raw.githubusercontent.com/u007/dockmetric/master/images/detail.png)

## usage

environment variables: 

* INFLUX_URL
* INFLUX_USER
* INFLUX_PASS

## build

```
env GOOS=linux GOARCH=amd64 go build -o examples/dockmetric main.go
```

## example usage

go into examples

edit docker-compose.yml and update:

* service.influx: INFLUXDB_ADMIN_PASSWORD 
* setup same password for influx.INFLUXDB_USER_PASSWORD and grafana.INFLUX_PASS and metrics.INFLUX_PASS
* setup grafana.GRAFANA_PASS with grafana password for admin login
* setup grafana.CONTAINERS and metrics.CONTAINERS with list of containers to monitor

then

```
cd examples
docker-compose build
docker-compose up

```
