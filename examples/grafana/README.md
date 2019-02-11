
To reset, delete:

```
rm -Rf grafana/custom
rm -Rf grafana/data
docker-compose build

```

to get plugin list

```
docker-compos exec grafana bash
grafana-cli plugins list-remote
```
