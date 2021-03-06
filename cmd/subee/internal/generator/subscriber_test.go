package generator_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/packages/packagestest"

	"github.com/wantedly/subee/cmd/subee/internal/generator"
)

func TestSubscriberGenerator(t *testing.T) {
	dir := filepath.Join(".", "testdata", "example.com", "a", "b")
	mods := []packagestest.Module{
		{Name: "example.com/a/b", Files: packagestest.MustCopyFileTree(dir)},
	}

	cases := []struct {
		test   string
		params generator.SubscriberParams
	}{
		{
			test:   "simple",
			params: generator.SubscriberParams{Name: "book"},
		},
		{
			test:   "with Message type",
			params: generator.SubscriberParams{Name: "book", Encoding: generator.MessageEncodingJSON, Package: generator.Package{Path: "./c"}, Message: "Book"},
		},
		{
			test:   "with Message type and alias",
			params: generator.SubscriberParams{Name: "author", Encoding: generator.MessageEncodingJSON, Package: generator.Package{Path: "./d"}, Message: "Author"},
		},
		{
			test:   "with Encoding protobuf",
			params: generator.SubscriberParams{Name: "book", Encoding: generator.MessageEncodingProtobuf, Package: generator.Package{Path: "./c"}, Message: "Book"},
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
					packagestest.TestAll(t, func(t *testing.T, exporter packagestest.Exporter) {
						exported := packagestest.Export(t, exporter, mods)
						defer exported.Cleanup()

						exported.Config.Mode = packages.LoadTypes
						if exporter.Name() == "GOPATH" {
							exported.Config.Dir = filepath.Join(exported.Config.Dir, "example.com", "a", "b")
						}

						gen := generator.NewSubscriberGenerator(exported.Config, ioutil.Discard)
						err := gen.Generate(context.Background(), &tc.params)
						if err != nil {
							t.Errorf("returned %+v, want nil", err)
						}

						err = filepath.Walk(exported.Config.Dir, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								t.Errorf("unexpected error: %v", err)
							}
							if info.IsDir() {
								return nil
							}
							relpath := strings.TrimPrefix(path, exported.Config.Dir)
							if !(strings.HasPrefix(relpath, "/pkg/") || strings.HasPrefix(relpath, "/cmd/")) {
								return nil
							}
							t.Run(relpath, func(t *testing.T) {
								data, err := exported.FileContents(path)
								if err != nil {
									t.Fatalf("retuend %v, want, nil", err)
								}
								cupaloy.SnapshotT(t, string(data))
							})
							return nil
						})

						if err != nil {
							t.Errorf("unexpected error: %v", err)
						}
					})
				})
			}
		})
	}
}
