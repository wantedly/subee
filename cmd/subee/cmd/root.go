package cmd

import "github.com/spf13/cobra"

func NewSubeeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "subee",
	}

	cmd.AddCommand(
		newGenerateCmd(),
		newStartCmd(),
	)

	return cmd
}
