package cmd

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/transform"
	"golang.org/x/tools/go/packages"

	"github.com/wantedly/subee/cmd/subee/internal/prefixer"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "start [NAME]",
		RunE: func(_ *cobra.Command, args []string) error {
			ctx := context.Background()

			wd, err := os.Getwd()
			if err != nil {
				return errors.WithStack(err)
			}

			dirs, err := filepath.Glob("./cmd/*-subscriber")
			if err != nil {
				return errors.WithStack(err)
			}
			for i, dir := range dirs {
				dirs[i] = "." + string(filepath.Separator) + dir
			}

			pkgCfg := &packages.Config{Context: ctx, Mode: packages.LoadSyntax, Dir: wd}
			pkgs, err := packages.Load(pkgCfg, dirs...)
			if err != nil {
				return errors.WithStack(err)
			}

			eg, ctx := errgroup.WithContext(ctx)
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			defer watchSignal(ctx, cancel)()

			for _, pkg := range pkgs {
				if len(pkg.Errors) > 0 || pkg.Name != "main" {
					continue
				}
				path := pkg.PkgPath
				eg.Go(func() error {
					cmd := exec.Command("go", "run", path)
					name := filepath.Base(path)
					color := determineColor(name)
					prefixer := prefixer.New(color(name).String)
					cmd.Stdout = transform.NewWriter(os.Stdout, prefixer)
					cmd.Stderr = transform.NewWriter(os.Stderr, prefixer)
					fmt.Fprintln(cmd.Stdout, "--> starting")
					defer fmt.Fprintln(cmd.Stdout, "<-- stopped")
					return errors.WithStack(execCommand(ctx, cmd))
				})
			}

			return errors.WithStack(eg.Wait())
		},
	}

	return cmd
}

func watchSignal(ctx context.Context, cb func()) (done func()) {
	sdCh := make(chan struct{})

	go func() {
		sigCh := make(chan os.Signal)
		signal.Notify(sigCh, os.Interrupt)
		for {
			select {
			case <-sigCh:
				cb()
			case <-sdCh:
				return
			}
		}
	}()

	return func() { close(sdCh) }
}

func execCommand(ctx context.Context, cmd *exec.Cmd) (err error) {
	errCh := make(chan error, 1)
	defer close(errCh)

	go func() { errCh <- cmd.Run() }()

	select {
	case <-ctx.Done():
		cmd.Process.Signal(os.Interrupt)
		err = <-errCh
	case err = <-errCh:
		// no-op
	}

	return
}

// https://github.com/wercker/stern/blob/1.10.0/stern/tail.go#L65-L81
var (
	colorList = []func(interface{}) aurora.Value{
		aurora.BrightRed,
		aurora.BrightGreen,
		aurora.BrightYellow,
		aurora.BrightBlue,
		aurora.BrightMagenta,
		aurora.BrightCyan,
	}
)

func determineColor(name string) func(interface{}) aurora.Value {
	hash := fnv.New32()
	hash.Write([]byte(name))
	idx := hash.Sum32() % uint32(len(colorList))

	return colorList[idx]
}
