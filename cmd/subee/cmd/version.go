package cmd

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	appName = "subee"
	version = "0.3.0"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "version",
		Short:         "Print the version information",
		Long:          "Print the version information",
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(_ *cobra.Command, _ []string) {
			buf := bytes.NewBufferString(appName + " " + version)
			buf.WriteString(" (")
			var meta []string
			for _, c := range []string{runtime.Version(), fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)} {
				if c != "" {
					meta = append(meta, c)
				}
			}
			buf.WriteString(strings.Join(meta, " "))
			buf.WriteString(")")
			fmt.Fprintln(os.Stdout, buf.String())
		},
	}
}
