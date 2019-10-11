package subee

import "context"

type (
	rawMessageContextKey  struct{}
	rawMessagesContextKey struct{}
)

func SetRawMessage(ctx context.Context, msg Message) context.Context {
	return context.WithValue(ctx, rawMessageContextKey{}, msg)
}

func GetRawMessage(ctx context.Context) Message {
	v := ctx.Value(rawMessageContextKey{})
	if v == nil {
		return nil
	}
	msg, ok := v.(Message)
	if !ok {
		return nil
	}
	return msg
}

func SetRawMessages(ctx context.Context, msgs []Message) context.Context {
	return context.WithValue(ctx, rawMessagesContextKey{}, msgs)
}

func GetRawMessages(ctx context.Context) []Message {
	v := ctx.Value(rawMessagesContextKey{})
	if v == nil {
		return nil
	}
	msgs, ok := v.([]Message)
	if !ok {
		return nil
	}
	return msgs
}
