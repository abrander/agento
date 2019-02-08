# agento
Client/server collecting near realtime metrics from Linux hosts. Uses influxdb as backend.



# development/debugging

## Agento
`DEBUG=* agento runonce`



## MySQL
`docker run --name agento-mysql -e MYSQL_ROOT_PASSWORD=agento -e MYSQL_USER=agento -e MYSQL_PASSWORD=agento -e MYSQL_DATABASE=mysql -p 3306:3306 -d mariadb:latest`

```
[probe.mysqltables]
interval = 1
dsn = "agento:agento@tcp(localhost)/mysql"
agent = "mysqltables"
```



## InfluxDB
`docker run --name agento-influxdb -p 8086:8086 -e INFLUXDB_DB=agento -e INFLUXDB_ADMIN_USER=agento -e INFLUXDB_ADMIN_PASSWORD=agento -d influxdb:latest`

```
[server.influxdb]
password = "agento"
url = "http://localhost:8086/"
username = "agento"
database = "agento"
retentionPolicy = "autogen"
```
