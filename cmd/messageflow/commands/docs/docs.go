package docs

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/denchenko/messageflow/internal/docs"
	"github.com/denchenko/messageflow/pkg/schema"
	"github.com/denchenko/messageflow/pkg/schema/target/d2"
	"github.com/spf13/cobra"
)

type Command struct {
	cmd *cobra.Command
}

// NewCommand creates a new gen-docs command
func NewCommand() *Command {
	c := &Command{}

	c.cmd = &cobra.Command{
		Use:   "gen-docs",
		Short: "Generate markdown documentation from AsyncAPI files",
		Long: `Generate comprehensive markdown documentation from AsyncAPI files.
Example:
  messageflow gen-docs --asyncapi-files asyncapi1.yaml,asyncapi2.yaml --output ./docs`,
		RunE: c.run,
	}

	c.cmd.Flags().String("asyncapi-files", "", "Paths to asyncapi files separated by comma")
	c.cmd.Flags().String("output", ".", "Output directory for generated documentation")
	c.cmd.Flags().String("title", "Message Flow", "Title of the documentation")

	err := c.cmd.MarkFlagRequired("asyncapi-files")
	if err != nil {
		log.Fatalf("error marking asyncapi-files flag as required: %v", err)
	}

	return c
}

// GetCommand returns the cobra command
func (c *Command) GetCommand() *cobra.Command {
	return c.cmd
}

func (c *Command) run(cmd *cobra.Command, _ []string) error {
	title, err := cmd.Flags().GetString("title")
	if err != nil {
		return fmt.Errorf("error getting title flag: %w", err)
	}

	asyncAPIFilesPath, err := cmd.Flags().GetString("asyncapi-files")
	if err != nil {
		return fmt.Errorf("error getting asyncapi-files flag: %w", err)
	}

	outputDir, err := cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("error getting output flag: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory %s: %w", outputDir, err)
	}

	ctx := context.Background()

	filePaths := strings.Split(asyncAPIFilesPath, ",")

	s, err := schema.Load(ctx, filePaths)
	if err != nil {
		return fmt.Errorf("error loading schema from files: %w", err)
	}

	d2Target, err := d2.NewTarget()
	if err != nil {
		return fmt.Errorf("error creating D2 target: %w", err)
	}

	if err := docs.Generate(ctx, s, d2Target, title, outputDir); err != nil {
		return fmt.Errorf("error generating documentation: %w", err)
	}

	fmt.Printf("Documentation generated successfully in: %s\n", outputDir)

	return nil
}
