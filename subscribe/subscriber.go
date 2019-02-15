package subscribe

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/api/option"

	"cloud.google.com/go/pubsub"
	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/wantedly/subee"
)

type subscriber struct {
	scfg               *subee.SubscribeConfig
	exBackoff          *backoff.ExponentialBackOff
	retry              Retry
	pubsubSubscription *pubsub.Subscription
	logger             subee.Logger
	sh                 subee.StatsHandler
	pubsubClientOption []option.ClientOption
}

// Retry is custom type for retry operation
type Retry func(c context.Context, operation func() error) error

// CreateSubscriber returns Subscriber implementation.
func CreateSubscriber(ctx context.Context, projectID, subscriptionID string, opts ...Option) (subee.Subscriber, error) {
	sub := &subscriber{
		exBackoff: backoff.NewExponentialBackOff(),
	}
	sub.retry = sub.retryOperation

	for _, opt := range opts {
		opt(sub)
	}

	c, err := createPubSubClient(ctx, projectID, sub.pubsubClientOption...)
	if err != nil {
		return nil, errors.Wrap(err, "faild to create Google Cloud Pub/Sub client")
	}

	pubsubSub, err := createPubSubSubscription(ctx, subscriptionID, c)
	if err != nil {
		return nil, errors.Wrap(err, "faild to create Google Cloud Pub/Sub subscription")
	}
	sub.pubsubSubscription = pubsubSub

	return sub, nil
}

// createSubscription returns *pubsub.Subscription instance
func createPubSubSubscription(ctx context.Context, subscriptionID string, c *pubsub.Client) (*pubsub.Subscription, error) {
	pubsubSub := c.Subscription(subscriptionID)

	ok, err := pubsubSub.Exists(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "faild to reports whether the subscription exists on the server")
	}

	if !ok {
		return nil, errors.New("failed to get pub/sub subscription. Check subscription presence")
	}

	return pubsubSub, nil
}

// Subscribe acquires a message from Google Clound Pubs/Sub and sends messages to channel.
func (s *subscriber) Subscribe(ctx context.Context, msgCh chan *subee.BufferedMessage) (err error) {
	s.logger, s.sh, s.scfg = subee.GetLogger(ctx), subee.GetStatsHandler(ctx), subee.GetSubscribeConfig(ctx)

	defer func() {
		close(msgCh)

		if pncErr := recover(); pncErr != nil {
			err = fmt.Errorf("panic on subscribing process: %#v", pncErr)
		}
	}()

	s.logger.Print("start subscribing")
	err = s.subscribe(ctx, msgCh)
	s.logger.Print("stop subscribing")
	return
}

// the receive type subscription model to retrieve messages.
func (s *subscriber) subscribe(ctx context.Context, msgCh chan *subee.BufferedMessage) error {
	for {
		select {
		// leave this loop when interrupted by os signal
		case <-ctx.Done():
			s.logger.Print("finish subscribing process, because of interrupted by os signal")
			return nil
		default:
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			ctx = s.sh.TagProcess(ctx, &subee.BeginTag{})
			ctx = s.sh.TagProcess(ctx, &subee.ReceiveBeginTag{})

			beginTime := time.Now()

			s.logger.Print("start receiving messages")

			var msgs []subee.Message

			operation := func() (err error) {
				msgs, err = s.receiveMessages(ctx)
				if err == nil {
					s.receiveEndHandleProcess(ctx, beginTime, nil)
				}

				return err
			}

			err := s.retry(ctx, operation)

			s.logger.Printf("stop receiving messages: received %d messages", len(msgs))

			if err != nil {
				s.receiveEndHandleProcess(ctx, beginTime, err)
				s.endHandleProcess(ctx, beginTime)

				return errors.WithStack(err)
			}

			if len(msgs) > 0 {
				ctx = s.sh.TagProcess(ctx, &subee.EnqueueTag{})
				ctx = context.WithValue(ctx, subee.EnqueueTimeContextKey{}, time.Now())
				ctx = context.WithValue(ctx, subee.BeginTimeContextKey{}, beginTime)

				msgCh <- &subee.BufferedMessage{
					Ctx:  ctx,
					Msgs: msgs,
				}

			} else {
				s.endHandleProcess(ctx, beginTime)
			}
		}
	}
}

func (s *subscriber) retryOperation(ctx context.Context, operation func() error) error {
	b := backoff.WithContext(s.exBackoff, ctx)
	err := backoff.Retry(operation, backoff.WithMaxRetries(b, uint64(s.scfg.MaxRetryCount)))
	return errors.WithStack(err)
}

func (s *subscriber) receiveEndHandleProcess(ctx context.Context, begin time.Time, err error) {
	s.sh.HandleProcess(ctx, &subee.ReceiveEnd{
		BeginTime: begin,
		EndTime:   time.Now(),
		Error:     err,
	})
}

func (s *subscriber) endHandleProcess(ctx context.Context, begin time.Time) {
	s.sh.HandleProcess(ctx, &subee.End{
		BeginTime: begin,
		EndTime:   time.Now(),
		MsgCount:  0,
	})
}

func (s *subscriber) receiveMessages(ctx context.Context) ([]subee.Message, error) {
	var mu sync.Mutex

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		time.Sleep(s.scfg.FlushInterval)
		cancel()
	}()

	buf := make(chan subee.Message, s.scfg.ChunkSize)
	err := s.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		mu.Lock()
		defer mu.Unlock()

		select {
		case buf <- &PubsubMessage{m}:
			// TODO(@hlts2): After adding the automatic ack function, erase it
			m.Ack()
		default:
			cancel()
			m.Nack()
		}
	})

	close(buf)

	if err != nil {
		return nil, errors.Wrap(err, "faild to receive message")
	}

	ms := make([]subee.Message, 0, len(buf))
	for m := range buf {
		ms = append(ms, m)
	}

	return ms, nil
}
