package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/wantedly/subee/cmd/subee/internal/generator"
)

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"g"},
	}

	cmd.AddCommand(
		newGenerateConsumerCmd(),
	)

	return cmd
}

func newGenerateConsumerCmd() *cobra.Command {
	var params generator.ConsumerParams
	params.Encoding = generator.MessageEncodingJSON

	cmd := &cobra.Command{
		Use:  "consumer NAME",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Name = args[0]
			err := generator.NewConsumerGenerator().Generate(context.Background(), &params)
			return errors.WithStack(err)
		},
	}

	cmd.Flags().StringVarP(&params.Package.Path, "package", "p", "", "")
	cmd.Flags().StringVarP(&params.Message, "message", "m", "", "")
	cmd.Flags().VarP(&params.Encoding, "encoding", "e", "")
	cmd.Flags().BoolVar(&params.Batch, "batch", false, "")

	return cmd
}
