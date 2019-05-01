package cmd

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"

	"github.com/wantedly/subee/cmd/subee/internal/generator"
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate a new code",
		Aliases: []string{"g"},
	}

	cmd.AddCommand(
		newGenerateSubscriberCmd(),
	)

	return cmd
}

func newGenerateSubscriberCmd() *cobra.Command {
	var params generator.SubscriberParams
	params.Encoding = generator.MessageEncodingJSON

	cmd := &cobra.Command{
		Use:   "subscriber NAME",
		Short: "Generate a new subscriber",
		Example: `
* Basic usage:

      subee g subscriber new-book


* Generate with a message deserializer:

      subee g subscriber -p ./path/to/model -m Book new-book


* Generate BatchConsumer:

      subee g subscriber -p ./path/to/model -m Book --batch new-book
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Name = args[0]
			wd, err := os.Getwd()
			if err != nil {
				return errors.WithStack(err)
			}
			cfg := packages.Config{Mode: packages.LoadTypes, Dir: wd}
			err = generator.NewSubscriberGenerator(&cfg).Generate(context.Background(), &params)
			return errors.WithStack(err)
		},
	}

	cmd.Flags().StringVarP(&params.Package.Path, "package", "p", "", "Path to package for a message type to use for deserializing messages")
	cmd.Flags().StringVarP(&params.Message, "message", "m", "", "Message type name to use for deserializing messages")
	cmd.Flags().VarP(&params.Encoding, "encoding", "e", "Message body encoding type, supports json or protobuf currently")
	cmd.Flags().BoolVar(&params.Batch, "batch", false, "Generate BatchConsumer")

	return cmd
}
