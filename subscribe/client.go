package subscribe

import (
	"context"

	"google.golang.org/api/option"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
)

func createPubSubClient(ctx context.Context, projectID string, ops ...option.ClientOption) (*pubsub.Client, error) {
	if len(projectID) == 0 {
		return nil, errors.New("missing google project id")
	}

	c, err := pubsub.NewClient(ctx, projectID, ops...)
	if err != nil {
		return nil, errors.Wrap(err, "faild to create pub/sub client")
	}

	return c, nil
}
