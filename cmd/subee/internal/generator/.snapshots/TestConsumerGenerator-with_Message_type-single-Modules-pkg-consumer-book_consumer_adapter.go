// Code generated by github.com/wantedly/subee/cmd/subee. DO NOT EDIT.

package consumer

import (
	"context"
	"encoding/json"

	"example.com/a/b/c"
	"github.com/pkg/errors"
	"github.com/wantedly/subee"
)

// NewBookConsumerAdapter created a consumer-adapter instance that converts incoming messages into c.Book.
func NewBookConsumerAdapter(consumer BookConsumer) subee.SingleMessageConsumer {
	return &bookConsumerAdapterImpl{consumer: consumer}
}

type bookConsumerAdapterImpl struct {
	consumer BookConsumer
}

func (a *bookConsumerAdapterImpl) Consume(ctx context.Context, m subee.Message) error {
	var err error
	obj := new(c.Book)
	err = json.Unmarshal(m.Data(), obj)
	if err != nil {
		return errors.WithStack(err)
	}
	err = a.consumer.Consume(ctx, obj)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

