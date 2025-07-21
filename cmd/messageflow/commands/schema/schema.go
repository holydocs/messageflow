package schema

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/holydocs/messageflow/pkg/messageflow"
	"github.com/holydocs/messageflow/pkg/schema"
	"github.com/holydocs/messageflow/pkg/schema/target/d2"
	"github.com/spf13/cobra"
)

type Command struct {
	cmd *cobra.Command
}

// NewCommand creates a new gen-schema command
func NewCommand() *Command {
	c := &Command{}

	c.cmd = &cobra.Command{
		Use:   "gen-schema",
		Short: "Generate schema from AsyncAPI files",
		Long: `Generate schema from AsyncAPI files and optionally format or render to output files.
		
Example:
  messageflow gen-schema --target d2 --render-to-file schema.svg --asyncapi-files asyncapi.yaml`,
		RunE: c.run,
	}

	c.cmd.Flags().String("target", "d2", "Target type (d2)")
	c.cmd.Flags().String("format-to-file", "", "Output file for the formatted schema")
	c.cmd.Flags().String("render-to-file", "", "Output file for the rendered diagram")
	c.cmd.Flags().String("asyncapi-files", "", "Paths to asyncapi files separated by comma")
	c.cmd.Flags().String("channel", "", "Channel")
	c.cmd.Flags().String("service", "", "Service")
	c.cmd.Flags().String("format-mode", "service_channels", "Format mode")
	c.cmd.Flags().Bool("omit-payloads", false, "Omit payloads")

	// Mark required flags
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

// run executes the gen-schema command
func (c *Command) run(cmd *cobra.Command, _ []string) error {
	targetType, err := cmd.Flags().GetString("target")
	if err != nil {
		return fmt.Errorf("error getting target flag: %w", err)
	}

	formatToFile, err := cmd.Flags().GetString("format-to-file")
	if err != nil {
		return fmt.Errorf("error getting format-to-file flag: %w", err)
	}

	renderToFile, err := cmd.Flags().GetString("render-to-file")
	if err != nil {
		return fmt.Errorf("error getting render-to-file flag: %w", err)
	}

	asyncAPIFilesPath, err := cmd.Flags().GetString("asyncapi-files")
	if err != nil {
		return fmt.Errorf("error getting asyncapi-files flag: %w", err)
	}

	channel, err := cmd.Flags().GetString("channel")
	if err != nil {
		return fmt.Errorf("error getting channel flag: %w", err)
	}

	service, err := cmd.Flags().GetString("service")
	if err != nil {
		return fmt.Errorf("error getting service flag: %w", err)
	}

	formatMode, err := cmd.Flags().GetString("format-mode")
	if err != nil {
		return fmt.Errorf("error getting format-mode flag: %w", err)
	}

	omitPayloads, err := cmd.Flags().GetBool("omit-payloads")
	if err != nil {
		return fmt.Errorf("error getting omit-payloads flag: %w", err)
	}

	// Validate that at least one output is specified
	if formatToFile == "" && renderToFile == "" {
		return errors.New("either --format-to-file or --render-to-file must be specified")
	}

	target, err := pickTarget(targetType)
	if err != nil {
		return fmt.Errorf("error picking target: %w", err)
	}

	targetCaps := target.Capabilities()

	if !targetCaps.Format {
		return errors.New("target doesn't support formatting")
	}

	if renderToFile != "" && !targetCaps.Render {
		return errors.New("target doesn't support render")
	}

	ctx := context.Background()

	filePaths := strings.Split(asyncAPIFilesPath, ",")

	s, err := schema.Load(ctx, filePaths)
	if err != nil {
		return fmt.Errorf("error loading schema from files: %w", err)
	}

	formatOpts := messageflow.FormatOptions{
		Mode:         messageflow.FormatMode(formatMode),
		Service:      service,
		Channel:      channel,
		OmitPayloads: omitPayloads,
	}

	fs, err := target.FormatSchema(ctx, s, formatOpts)
	if err != nil {
		return fmt.Errorf("error formatting schema: %w", err)
	}

	if formatToFile != "" {
		err = os.WriteFile(formatToFile, fs.Data, 0600)
		if err != nil {
			return fmt.Errorf("error writing to file %s: %w", formatToFile, err)
		}
		fmt.Printf("Formatted schema written to: %s\n", formatToFile)
	}

	if renderToFile != "" {
		diagram, err := target.RenderSchema(ctx, fs)
		if err != nil {
			return fmt.Errorf("error rendering schema: %w", err)
		}

		err = os.WriteFile(renderToFile, diagram, 0600)
		if err != nil {
			return fmt.Errorf("error writing to file %s: %w", renderToFile, err)
		}
		fmt.Printf("Rendered diagram written to: %s\n", renderToFile)
	}

	return nil
}

// pickTarget selects the appropriate target based on the target type
func pickTarget(targetType string) (messageflow.Target, error) {
	switch targetType {
	case "d2":
		return d2.NewTarget()
	default:
		return nil, fmt.Errorf("unknown target: %s", targetType)
	}
}
