package generator_test

import (
	"context"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/packages/packagestest"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/wantedly/subee/cmd/subee/internal/generator"
)

func TestConsumerGenerator(t *testing.T) {
	dir := filepath.Join(".", "testdata", "example.com", "a", "b")
	mods := []packagestest.Module{
		{Name: "example.com/a/b", Files: packagestest.MustCopyFileTree(dir)},
	}

	cases := []struct {
		test      string
		params    generator.ConsumerParams
		wantFiles []string
	}{
		{
			test:   "simple",
			params: generator.ConsumerParams{Name: "book"},
			wantFiles: []string{
				"cmd/book-worker/main.go",
				"cmd/book-worker/run.go",
				"pkg/consumer/book_consumer.go",
			},
		},
		{
			test:   "with Message type",
			params: generator.ConsumerParams{Name: "book", Encoding: generator.MessageEncodingJSON, Package: generator.Package{Path: "./c"}, Message: "Book"},
			wantFiles: []string{
				"cmd/book-worker/main.go",
				"cmd/book-worker/run.go",
				"pkg/consumer/book_consumer.go",
				"pkg/consumer/book_consumer_adapter.go",
			},
		},
		{
			test:   "with Message type and alias",
			params: generator.ConsumerParams{Name: "author", Encoding: generator.MessageEncodingJSON, Package: generator.Package{Path: "./d"}, Message: "Author"},
			wantFiles: []string{
				"cmd/author-worker/main.go",
				"cmd/author-worker/run.go",
				"pkg/consumer/author_consumer.go",
				"pkg/consumer/author_consumer_adapter.go",
			},
		},
		{
			test:   "with Encoding protobuf",
			params: generator.ConsumerParams{Name: "book", Encoding: generator.MessageEncodingProtobuf, Package: generator.Package{Path: "./c"}, Message: "Book"},
			wantFiles: []string{
				"cmd/book-worker/main.go",
				"cmd/book-worker/run.go",
				"pkg/consumer/book_consumer.go",
				"pkg/consumer/book_consumer_adapter.go",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.test, func(t *testing.T) {
			for _, b := range []bool{false, true} {
				tc.params.Batch = b
				name := "single"
				if b {
					name = "batch"
				}
				t.Run(name, func(t *testing.T) {
					tc := tc
					t.Parallel()
					packagestest.TestAll(t, func(t *testing.T, exporter packagestest.Exporter) {
						exported := packagestest.Export(t, exporter, mods)
						defer exported.Cleanup()

						exported.Config.Mode = packages.LoadTypes
						if exporter.Name() == "GOPATH" {
							exported.Config.Dir = filepath.Join(exported.Config.Dir, "example.com", "a", "b")
						}

						gen := generator.NewConsumerGenerator(exported.Config)
						err := gen.Generate(context.Background(), &tc.params)
						if err != nil {
							t.Errorf("returned %+v, want nil", err)
						}

						for _, f := range tc.wantFiles {
							f := f
							t.Run(f, func(t *testing.T) {
								data, err := exported.FileContents(filepath.Join(exported.Config.Dir, f))
								if err != nil {
									t.Fatalf("retuend %v, want, nil", err)
								}
								cupaloy.SnapshotT(t, string(data))
							})
						}
					})
				})
			}
		})
	}
}
