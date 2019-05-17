# Your first Google Cloud Pub/Sub app

## Files

```
├── cmd
│   └── event-subscriber
│       ├── main.go
│       └── run.go    # entrypoint
├── go.mod
├── go.sum
└── pkg
    ├── consumer
    │   ├── event_consumer.go          # message handler implementation(usually contains business logic)
    │   └── event_consumer_adapter.go  # auto-generated type converter
    ├── model
    │   └── event.go    # message type
    └── util
        └── publish.go  # utility for creating Pub/Sub resources and publish messages
```

## Requirements

- Docker
- Golang (>= 1.11)
- subee CLI (recommended)

## Running


- Start Cloud Pub/Sub emulator
    - ```
      $ docker run \
        -p 8085:8085 \
        --name subee-examples \
        -e CLOUDSDK_CORE_PROJECT=subee-examples \
        -d \
        google/cloud-sdk \
        gcloud beta emulators pubsub start --host-port=0.0.0.0:8085
      ```
- Set `PUBSUB_EMULATOR_HOST`
    - ```
      $ eval $(docker exec subee-examples gcloud beta emulators pubsub env-init)
      ```
- Start subscriber and publish messages
    - ```bash
      $ subee start
      event-subscriber ▸ starting
      event-subscriber 2019/05/17 14:11:37 Start Pub/Sub worker
      event-subscriber 2019/05/17 14:11:37 Start process
      event-subscriber 2019/05/17 14:11:37 Start subscribing messages
      event-subscriber 2019/05/17 14:11:38 Start consuming 1 messages
      event-subscriber 2019/05/17 14:11:38 Received event, created_at: 1558069898274607000
      event-subscriber 2019/05/17 14:11:38 Finish consuming 1 messages
      event-subscriber 2019/05/17 14:11:39 Start consuming 1 messages
      event-subscriber 2019/05/17 14:11:39 Received event, created_at: 1558069899274530000
      event-subscriber 2019/05/17 14:11:39 Finish consuming 1 messages
      event-subscriber 2019/05/17 14:11:40 Start consuming 1 messages
      event-subscriber 2019/05/17 14:11:40 Received event, created_at: 1558069900275306000
      event-subscriber 2019/05/17 14:11:40 Finish consuming 1 messages
      event-subscriber 2019/05/17 14:11:41 Start consuming 1 messages
      event-subscriber 2019/05/17 14:11:41 Received event, created_at: 1558069901274765000
      event-subscriber 2019/05/17 14:11:41 Finish consuming 1 messages
      ```
