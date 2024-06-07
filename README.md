# RRD Service
Dmitrii Neeman (07.06.2024) Golang 1.22.2 [spent ~8 hours]

## Used libraries
- `github.com/aerospike/aerospike-client-go/v7` - for saving and processing data in aerospike databse.
- `github.com/gorilla/mux` - for managing http router.
- `github.com/ilyakaznacheev/cleanenv` - for loading env variables.
- `github.com/steinfletcher/apitest` - for http api tests.
- `github.com/stretchr/testify` - for tests.

## Testing
```bash
./ci/run_tests.sh
```

## Building
```bash
go build cmd/main/rrd.go
```

## Configuration
Service retrieves configuration parameters form ENV. If ENV is empty it uses default values:
- `LOG_LEVEL` - logger level (default: info)
- `HTTP_PORT` - https server port (default: 8080)
- `STORAGE_CAP` - maximum metric capacity (default: 1000)
- `STORAGE_HOST` - aerospike database host (default: localhost)
- `STORAGE_PORT` - aerospike database port (default: 3000)
- `STORAGE_NAMESPACE` - aerospike database namespace (default: test)

## Running
```bash
./rrd
```

## Run in container.
### Build
```bash
docker build -t rrd-service .
```
### Run
```bash
docker network create rrd-network
docker run --name as-test --net rrd-network -p 127.0.0.1:3000:3000 aerospike:ee-7.1.0.0_2
docker run --name rrd-test --net rrd-network -e STORAGE_HOST=as-test -p 127.0.0.1:8080:8080 rrd-service
```

## Project structure
- `ci` - test script for ci integration.
- `cmd` - running application.
- `internal` - application logic.
    - `adaptors` - adaptors for storage.
        - `storage` - database logic for aerospike storage.
    - `config` - parsing and loading config params from ENV.
    - `httpsrv` - http server.
        - `handlers` - http handlers.
    - `models` - contains entities that are used by the application.
    - `rrd` - application logic.
    - `app.go` - services initialization, starting server.
- `udf` - user defined function for aerospike.

## Usage
Take a look at `swagger.yaml`

### Put metric
`[PUT] /metrics`
- Request
```json
  {
    "timestamp": 1717745157997559,
    "metric_value": 11.5
  }
```

### Get metrics
`[GET] /metrics?start=0&end=1717745157997559`
- Response
```json
  [
    {"timestamp":1717745157997559,"metric_value":11.5}
  ]
```
  
## Notice
- I've spent a lot of time, reading aerospike documentation and gathering information on forums, that's why I spent ~8 hours.
- We set ttl for records through `aerospike.NewWritePolicy(0, aerospike.TTLDontExpire)`
(Also we can set this parameter globally in config file)
- All record limitation logic is implemented in storage, 
because if we decide to change storage, we'll need to rewrite only this part.
- Eviction is implemented using udf. We find the latest record and delete it if we reach the cap.
(Maybe we can implement all eviction logic in udf, but I haven't got enough time to research)
- Counters are stored in a database, so after restart we have the current value. 
(This logic can be also implemented by counting all the records on start, but counting records never was a fast operation.)
- I didn't use any validation library because validations here are basic.
- I didn't undertake phrase in requirements `Support multiple metrics`. 
If it means that we want to save metrics for different service (sources), 
then we can easily achieve this by adding Source fields to our `models.Record`. 
If it means different metric types, I've made `MetricValue` of type `any`, so we can save any value there.

  