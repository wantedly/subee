// Code generated by github.com/wantedly/subee/cmd/subee. DO NOT EDIT.

package consumer

import (
	"context"
	"encoding/json"

	"example.com/a/b/c"
	"github.com/pkg/errors"
	"github.com/wantedly/subee"
)

// NewBookBatchConsumerAdapter created a consumer-adapter instance that converts incoming messages into c.Book.
func NewBookBatchConsumerAdapter(consumer BookBatchConsumer) subee.BatchConsumer {
	return &bookBatchConsumerAdapterImpl{consumer: consumer}
}

type bookBatchConsumerAdapterImpl struct {
	consumer BookBatchConsumer
}

func (a *bookBatchConsumerAdapterImpl) BatchConsume(ctx context.Context, ms []subee.Message) error {
	var err error
	objs := make([]*c.Book, len(ms))
	for i, m := range ms {
		obj := new(c.Book)
		err = json.Unmarshal(m.Data(), obj)
		if err != nil {
			return errors.WithStack(err)
		}
		objs[i] = obj
	}
	err = a.consumer.BatchConsume(ctx, objs)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

