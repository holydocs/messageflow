package docs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/holydocs/messageflow/internal/docs"
	"github.com/holydocs/messageflow/pkg/schema"
	"github.com/holydocs/messageflow/pkg/schema/target/d2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

	c.cmd.Flags().String("dir", "", "Path to dir to scan asyncapi files automatically")
	c.cmd.Flags().String("asyncapi-files", "", "Paths to asyncapi files separated by comma")
	c.cmd.Flags().String("output", ".", "Output directory for generated documentation")
	c.cmd.Flags().String("title", "Message Flow", "Title of the documentation")

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

	asyncAPIFilesPaths, err := getAsyncAPIFilesPaths(cmd)
	if err != nil {
		return fmt.Errorf("error getting asyncapi files paths: %w", err)
	}

	outputDir, err := cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("error getting output flag: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory %s: %w", outputDir, err)
	}

	ctx := context.Background()

	s, err := schema.Load(ctx, asyncAPIFilesPaths)
	if err != nil {
		return fmt.Errorf("error loading schema from files: %w", err)
	}

	d2Target, err := d2.NewTarget()
	if err != nil {
		return fmt.Errorf("error creating D2 target: %w", err)
	}

	newChangelog, err := docs.Generate(ctx, s, d2Target, title, outputDir)
	if err != nil {
		return fmt.Errorf("error generating documentation: %w", err)
	}

	fmt.Printf("Documentation generated successfully in: %s\n", outputDir)

	if newChangelog != nil && len(newChangelog.Changes) > 0 {
		fmt.Printf("\nNew Changes Detected:\n")
		for _, change := range newChangelog.Changes {
			fmt.Printf("• %s %s: %s\n", change.Type, change.Category, change.Details)
			if change.Diff != "" {
				fmt.Println(change.Diff)
			}
		}
	}

	return nil
}

func getAsyncAPIFilesPaths(cmd *cobra.Command) ([]string, error) {
	asyncAPIFilesPath, err := cmd.Flags().GetString("asyncapi-files")
	if err != nil {
		return nil, fmt.Errorf("error getting asyncapi-files flag: %w", err)
	}

	if asyncAPIFilesPath != "" {
		return strings.Split(asyncAPIFilesPath, ","), nil
	}

	asyncAPIFilesDir, err := cmd.Flags().GetString("dir")
	if err != nil {
		return nil, fmt.Errorf("error getting dir flag: %w", err)
	}

	if asyncAPIFilesDir == "" {
		return nil, errors.New("provide either asyncapi-files or dir")
	}

	return asyncAPIFilesFromDir(asyncAPIFilesDir)
}

func asyncAPIFilesFromDir(dir string) ([]string, error) {
	fmt.Println("Scanning directory for AsyncAPI files:", dir)

	var asyncAPIFiles []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		var yamlDoc map[string]interface{}
		if err := yaml.Unmarshal(content, &yamlDoc); err != nil {
			return fmt.Errorf("error unmarshalling yaml file %s: %w", path, err)
		}

		if _, hasAsyncAPI := yamlDoc["asyncapi"]; hasAsyncAPI {
			asyncAPIFiles = append(asyncAPIFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", dir, err)
	}

	if len(asyncAPIFiles) == 0 {
		return nil, fmt.Errorf("no AsyncAPI specification files found in directory %s", dir)
	}

	fmt.Println("Found AsyncAPI files:", asyncAPIFiles)

	return asyncAPIFiles, nil
}
