# Your first Google Cloud Pub/Sub app

## Files

- [main.go](main.go) - example source code
- [docker-compose.yml](docker-compose.yml) - tool for defining and running multi container Docker applications, contains Golang, Google Cloud Pub/Sub Emulator
- [go.mod](go.mod) - Go modules dependencies
- [go.sum](go.sum) - Go modules checksums

You can find more information about `go.mod` and `go.sum` at [Go wiki](https://github.com/golang/go/wiki/Modules)

## Requirements

To run this example you will need Docker and docker-compose installed. See installation guide at https://docs.docker.com/compose/install/

## Running

```bash
> docker-compose up
[a lot of Cloud Pub/Sub logs...]
golang_1     | 2019/02/26 07:56:39 Start Pub/Sub worker
golang_1     | 2019/02/26 07:56:39 Start consume process
golang_1     | 2019/02/26 07:56:40 Received event, created_at: 1551167800870017994
golang_1     | 2019/02/26 07:56:41 Received event, created_at: 1551167801870989467
golang_1     | 2019/02/26 07:56:42 Received event, created_at: 1551167802869335604
golang_1     | 2019/02/26 07:56:43 Received event, created_at: 1551167803872802935
golang_1     | 2019/02/26 07:56:44 Received event, created_at: 1551167804871915174
golang_1     | 2019/02/26 07:56:45 Received event, created_at: 1551167805870712427
```
