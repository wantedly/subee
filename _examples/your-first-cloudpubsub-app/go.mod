module your-first-cloudpubsub-app

go 1.13

require (
	cloud.google.com/go/pubsub v1.0.1
	github.com/pkg/errors v0.8.1
	github.com/wantedly/subee v0.4.0
)

replace github.com/wantedly/subee => ../..
