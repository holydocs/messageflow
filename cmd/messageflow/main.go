package main

import (
	"fmt"
	"os"

	"github.com/denchenko/messageflow/cmd/messageflow/commands/schema"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "messageflow",
		Short: "MessageFlow - AsyncAPI schema processing tool",
		Long:  `MessageFlow is a tool for generating schemas/docs from AsyncAPI schemas.`}

	rootCmd.AddCommand(schema.NewCommand().GetCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
