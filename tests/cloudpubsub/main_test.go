package cloudpubsub_test

import (
	"context"
	"io/ioutil"
	"log"
	"reflect"
	"sync"
	"testing"

	"cloud.google.com/go/pubsub"

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
	in := [][]byte{
		[]byte("foo"),
		[]byte("var"),
	}

	type TestCase struct {
		test string
		want *pubsub.Message
	}

	type Result struct {
		msg subee.Message
		tc  TestCase
	}

	cases := []TestCase{
		{
			test: "acked when no error",
			want: &pubsub.Message{
				Data: in[0],
			},
		},
		{
			test: "nacked when errored",
			want: &pubsub.Message{
				Data: in[1],
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// const subscriptionID = "single_message_consumer"
	const subscriptionID = "test1"

	publisher, err := createPublisher(ctx, projectID, topicID, subscriptionID)
	if err != nil {
		t.Fatalf("createPublisher returned error: %v", err)
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	resultCh := make(chan *Result, len(cases))

	subscriber, err := cloudpubsub.CreateSubscriber(ctx, projectID, subscriptionID)
	if err != nil {
		t.Fatalf("CreateSubscriber returned an error: %v", err)
	}

	var cur int
	consumer := subee.SingleMessageConsumerFunc(func(ctx context.Context, msg subee.Message) error {
		r := &Result{
			tc:  cases[cur],
			msg: msg,
		}

		resultCh <- r

		cur++

		if cur == len(cases) {
			close(resultCh)
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

	for r := range resultCh {
		t.Run(r.tc.test, func(t *testing.T) {
			got, want := r.msg, r.tc.want

			if got, want := got.Data(), want.Data; !reflect.DeepEqual(got, want) {
				t.Errorf("Data is %q, want: %q", got, want)
			}
		})
	}
}
