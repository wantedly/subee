// Code generated by github.com/wantedly/subee/cmd/subee. DO NOT EDIT.

package consumer

import (
	"context"
	"encoding/json"

	e "example.com/a/b/d"
	"github.com/pkg/errors"
	"github.com/wantedly/subee"
)

// NewAuthorConsumerAdapter created a consumer-adapter instance that converts incoming messages into e.Author.
func NewAuthorConsumerAdapter(consumer AuthorConsumer) subee.SingleMessageConsumer {
	return &authorConsumerAdapterImpl{consumer: consumer}
}

type authorConsumerAdapterImpl struct {
	consumer AuthorConsumer
}

func (a *authorConsumerAdapterImpl) Consume(ctx context.Context, ms []subee.Message) error {
	var err error
	objs := make([]*e.Author, len(ms))
	for i, m := range ms {
		obj := new(e.Author)
		err = json.Unmarshal(m.Data(), obj)
		if err != nil {
			return errors.WithStack(err)
		}
		objs[i] = obj
	}
	err = a.consumer.Consume(ctx, objs)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

