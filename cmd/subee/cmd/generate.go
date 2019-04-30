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
		Use:  "subscriber NAME",
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

	cmd.Flags().StringVarP(&params.Package.Path, "package", "p", "", "")
	cmd.Flags().StringVarP(&params.Message, "message", "m", "", "")
	cmd.Flags().VarP(&params.Encoding, "encoding", "e", "")
	cmd.Flags().BoolVar(&params.Batch, "batch", false, "")

	return cmd
}
