# Your first Google Cloud Pub/Sub app

## Files

- [main.go](main.go) - example source code
- [docker-compose.yml](docker-compose.yml) - tool for defining and running multi container Docker applications, contains Golang, Google Cloud Pub/Sub Emulator
- [go.mod](go.mod) - Go modules dependencies, [Go wiki](https://github.com/golang/go/wiki/Modules)
- [go.sum](go.sum) - Go modules checksums

## Requirements

To run this example you will need Docker and docker-compose installed.

See installation guide at https://docs.docker.com/compose/install/

## Running

`bash
> docker-compose up
[a lot of Cloud Pub/Sub logs...]
golang_1              | {"level":"info","ts":1551161785.525428,"caller":"subee@v0.1.0/engine.go:46","msg":"Start Pub/Sub worker"}
golang_1              | {"level":"info","ts":1551161785.5255408,"caller":"subee@v0.1.0/engine.go:126","msg":"Start consume process"}
golang_1              | {"level":"info","ts":1551161786.5839193,"caller":"zap/zap_interceptor.go:55","msg":"Start consume message.","message_count":1}
golang_1              | {"level":"info","ts":1551161786.584311,"caller":"app/main.go:124","msg":"Received event","created_at":1551161786513040831}
golang_1              | {"level":"info","ts":1551161786.584385,"caller":"zap/zap_interceptor.go:62","msg":"End consume message.","message_count":1,"time":0.000357034}
golang_1              | {"level":"info","ts":1551161787.546799,"caller":"zap/zap_interceptor.go:55","msg":"Start consume message.","message_count":1}
golang_1              | {"level":"info","ts":1551161787.5468953,"caller":"app/main.go:124","msg":"Received event","created_at":1551161787512175761}
golang_1              | {"level":"info","ts":1551161787.5469103,"caller":"zap/zap_interceptor.go:62","msg":"End consume message.","message_count":1,"time":0.000029096}
golang_1              | {"level":"info","ts":1551161788.5289135,"caller":"zap/zap_interceptor.go:55","msg":"Start consume message.","message_count":1}
golang_1              | {"level":"info","ts":1551161788.529226,"caller":"app/main.go:124","msg":"Received event","created_at":1551161788510769372}
golang_1              | {"level":"info","ts":1551161788.5292633,"caller":"zap/zap_interceptor.go:62","msg":"End consume message.","message_count":1,"time":0.000262977}
golang_1              | {"level":"info","ts":1551161789.551853,"caller":"zap/zap_interceptor.go:55","msg":"Start consume message.","message_count":1}
golang_1              | {"level":"info","ts":1551161789.5519247,"caller":"app/main.go:124","msg":"Received event","created_at":1551161789512000811}
golang_1              | {"level":"info","ts":1551161789.5519412,"caller":"zap/zap_interceptor.go:62","msg":"End consume message.","message_count":1,"time":0.00003399}
golang_1              | {"level":"info","ts":1551161790.5700421,"caller":"zap/zap_interceptor.go:55","msg":"Start consume message.","message_count":1}
golang_1              | {"level":"info","ts":1551161790.5705235,"caller":"app/main.go:124","msg":"Received event","created_at":1551161790511854820}
golang_1              | {"level":"info","ts":1551161790.5706062,"caller":"zap/zap_interceptor.go:62","msg":"End consume message.","message_count":1,"time":0.000104842}
golang_1              | {"level":"info","ts":1551161791.5361006,"caller":"zap/zap_interceptor.go:55","msg":"Start consume message.","message_count":1}
golang_1              | {"level":"info","ts":1551161791.536451,"caller":"app/main.go:124","msg":"Received event","created_at":1551161791510733309}
golang_1              | {"level":"info","ts":1551161791.5365493,"caller":"zap/zap_interceptor.go:62","msg":"End consume message.","message_count":1,"time":0.000113981}
golang_1              | {"level":"info","ts":1551161792.551914,"caller":"zap/zap_interceptor.go:55","msg":"Start consume message.","message_count":1}
```
