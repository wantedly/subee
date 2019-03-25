package cloudpubsub_test

import (
	"context"
	"io/ioutil"
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wantedly/subee"
	"github.com/wantedly/subee/subscribers/cloudpubsub"
)

const (
	// ProjectID is google cloud project id for test.
	projectID = "project_id_for_subee_test"

	// topicID is pub/sub topic id for test.
	topicID = "topic_id_for_subee_test"
)

func TestEngineWithConsumer(t *testing.T) {
	in := createInputData(256)

	ctx, cancel := context.WithCancel(context.Background())

	const subscriptionID = "subscription_id_for_consumer"

	publisher, err := createPublisher(ctx, projectID, topicID, subscriptionID)
	if err != nil {
		t.Fatalf("createPublisher returned error: %v", err)
	}

	subscriber, err := cloudpubsub.CreateSubscriber(ctx, projectID, subscriptionID)
	if err != nil {
		t.Fatalf("CreateSubscriber returned an error: %v", err)
	}

	var cnt int64

	consumer := subee.ConsumerFunc(func(ctx context.Context, msg subee.Message) error {
		atomic.AddInt64(&cnt, 1)
		return nil
	})

	engine := subee.New(
		subscriber,
		consumer,
		subee.WithLogger(log.New(ioutil.Discard, "", 0)),
	)

	go func() {
		publisher.publish(ctx, in)
		time.Sleep(1 * time.Second)
		cancel()
	}()

	err = engine.Start(ctx)
	if err != nil {
		t.Errorf("Start returned an error: %v", err)
	}

	got := int(atomic.LoadInt64(&cnt))
	if want := len(in); want != got {
		t.Errorf("want %d, got: %v", want, got)
	}
}

func TestEngineWithBatchConsumer(t *testing.T) {
	in := createInputData(256)

	ctx, cancel := context.WithCancel(context.Background())

	const subscriptionID = "subscription_id_for_batch_consumer"

	publisher, err := createPublisher(ctx, projectID, topicID, subscriptionID)
	if err != nil {
		t.Fatalf("createPublisher returned error: %v", err)
	}

	subscriber, err := cloudpubsub.CreateSubscriber(ctx, projectID, subscriptionID)
	if err != nil {
		t.Fatalf("CreateSubscriber returned an error: %v", err)
	}

	var cnt int64

	consumer := subee.BatchConsumerFunc(func(ctx context.Context, msgs []subee.Message) error {
		atomic.AddInt64(&cnt, int64(len(msgs)))
		return nil
	})

	engine := subee.NewBatch(
		subscriber,
		consumer,
		subee.WithChunkSize(3),
		subee.WithFlushInterval(4*time.Millisecond),
		subee.WithLogger(log.New(ioutil.Discard, "", 0)),
	)

	go func() {
		publisher.publish(ctx, in)
		time.Sleep(1 * time.Second)
		cancel()
	}()

	err = engine.Start(ctx)
	if err != nil {
		t.Errorf("Start returned an error: %v", err)
	}

	got := int(atomic.LoadInt64(&cnt))
	if want := len(in); want != got {
		t.Errorf("want %d, got: %v", want, got)
	}
}
