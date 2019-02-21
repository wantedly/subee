package cloudpubsub_test

import (
	"context"
	"io/ioutil"
	"log"
	"sync"
	"testing"

	"github.com/wantedly/subee"
	"github.com/wantedly/subee/subscribers/cloudpubsub"
)

const (
	// ProjectID is google cloud project id for test.
	projectID = "project_id_for_subee_test"

	// topicID is pub/sub topic id for test.
	topicID = "topic_id_for_subee_test"
)

func TestEngineWithSingleMessageConsumer(t *testing.T) {
	in := make([][]byte, 10)
	for i := 0; i < len(in); i++ {
		in[i] = []byte{byte(i)}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const subscriptionID = "subscription_id_for_single_message_consumer"

	publisher, err := createPublisher(ctx, projectID, topicID, subscriptionID)
	if err != nil {
		t.Fatalf("createPublisher returned error: %v", err)
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	subscriber, err := cloudpubsub.CreateSubscriber(ctx, projectID, subscriptionID)
	if err != nil {
		t.Fatalf("CreateSubscriber returned an error: %v", err)
	}

	msgCh := make(chan subee.Message)

	var cur int
	consumer := subee.SingleMessageConsumerFunc(func(ctx context.Context, msg subee.Message) error {
		msgCh <- msg

		cur++

		if cur == len(in) {
			close(msgCh)
			cancel()
		}

		return nil
	})

	engine := subee.NewWithSingleMessageConsumer(
		subscriber,
		consumer,
		subee.WithLogger(log.New(ioutil.Discard, "", 0)),
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := engine.Start(ctx)
		if err != nil {
			t.Errorf("Start returned an error: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		publisher.publish(ctx, in)
	}()

	var seen [10]bool
	for m := range msgCh {
		seen[m.Data()[0]] = true
	}

	for i, saw := range seen {
		if !saw {
			t.Errorf("did not see message #%d", i)
		}
	}
}
