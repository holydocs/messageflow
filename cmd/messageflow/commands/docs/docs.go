package docs

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/denchenko/messageflow"
	"github.com/denchenko/messageflow/source/asyncapi"
	"github.com/denchenko/messageflow/target/d2"
	"github.com/spf13/cobra"
)

//go:embed templates/readme.tmpl
var readmeTemplateFS embed.FS

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
	schemas := make([]messageflow.Schema, 0, len(filePaths))

	for _, filePath := range filePaths {
		trimmedPath := strings.TrimSpace(filePath)
		s, err := asyncapi.NewSource(trimmedPath)
		if err != nil {
			return fmt.Errorf("error creating schema source from %s: %w", trimmedPath, err)
		}

		schema, err := s.ExtractSchema(ctx)
		if err != nil {
			return fmt.Errorf("error extracting schema from %s: %w", trimmedPath, err)
		}

		schemas = append(schemas, schema)
	}

	mergedSchema := messageflow.MergeSchemas(schemas...)

	d2Target, err := d2.NewTarget()
	if err != nil {
		return fmt.Errorf("error creating D2 target: %w", err)
	}

	title, err := cmd.Flags().GetString("title")
	if err != nil {
		return fmt.Errorf("error getting title flag: %w", err)
	}

	if err := c.generateDocumentation(ctx, mergedSchema, d2Target, title, outputDir); err != nil {
		return fmt.Errorf("error generating documentation: %w", err)
	}

	fmt.Printf("Documentation generated successfully in: %s\n", outputDir)
	return nil
}

func (c *Command) generateDocumentation(
	ctx context.Context,
	schema messageflow.Schema,
	target messageflow.Target,
	title, outputDir string,
) error {
	contextDiagram, err := c.generateContextDiagram(ctx, schema, target)
	if err != nil {
		return fmt.Errorf("error generating context diagram: %w", err)
	}

	serviceDiagrams := make(map[string][]byte)
	for _, service := range schema.Services {
		diagram, err := c.generateServiceChannelsDiagram(ctx, schema, target, service.Name)
		if err != nil {
			return fmt.Errorf("error generating service channels diagram for %s: %w", service.Name, err)
		}
		serviceDiagrams[service.Name] = diagram
	}

	channelDiagrams := make(map[string][]byte)
	channels := c.extractUniqueChannels(schema)
	for _, channel := range channels {
		diagram, err := c.generateChannelServicesDiagram(ctx, schema, target, channel)
		if err != nil {
			return fmt.Errorf("error generating channel services diagram for %s: %w", channel, err)
		}
		channelDiagrams[channel] = diagram
	}

	if err := c.writeDiagramFiles(outputDir, contextDiagram, serviceDiagrams, channelDiagrams); err != nil {
		return fmt.Errorf("error writing diagram files: %w", err)
	}

	readmeContent, err := c.createREADMEContent(schema, title)
	if err != nil {
		return fmt.Errorf("error creating README content: %w", err)
	}

	readmePath := filepath.Join(outputDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("error writing README.md: %w", err)
	}

	return nil
}

func (c *Command) generateContextDiagram(
	ctx context.Context,
	schema messageflow.Schema,
	target messageflow.Target,
) ([]byte, error) {
	formatOpts := messageflow.FormatOptions{
		Mode: messageflow.FormatModeContextServices,
	}

	formattedSchema, err := target.FormatSchema(ctx, schema, formatOpts)
	if err != nil {
		return nil, fmt.Errorf("error formatting context schema: %w", err)
	}

	diagram, err := target.RenderSchema(ctx, formattedSchema)
	if err != nil {
		return nil, fmt.Errorf("error rendering context diagram: %w", err)
	}

	return diagram, nil
}

func (c *Command) generateServiceChannelsDiagram(
	ctx context.Context,
	schema messageflow.Schema,
	target messageflow.Target,
	serviceName string,
) ([]byte, error) {
	formatOpts := messageflow.FormatOptions{
		Mode:    messageflow.FormatModeServiceChannels,
		Service: serviceName,
	}

	formattedSchema, err := target.FormatSchema(ctx, schema, formatOpts)
	if err != nil {
		return nil, fmt.Errorf("error formatting service channels schema: %w", err)
	}

	diagram, err := target.RenderSchema(ctx, formattedSchema)
	if err != nil {
		return nil, fmt.Errorf("error rendering service channels diagram: %w", err)
	}

	return diagram, nil
}

func (c *Command) generateChannelServicesDiagram(
	ctx context.Context,
	schema messageflow.Schema,
	target messageflow.Target,
	channel string,
) ([]byte, error) {
	formatOpts := messageflow.FormatOptions{
		Mode:    messageflow.FormatModeChannelServices,
		Channel: channel,
	}

	formattedSchema, err := target.FormatSchema(ctx, schema, formatOpts)
	if err != nil {
		return nil, fmt.Errorf("error formatting channel services schema: %w", err)
	}

	diagram, err := target.RenderSchema(ctx, formattedSchema)
	if err != nil {
		return nil, fmt.Errorf("error rendering channel services diagram: %w", err)
	}

	return diagram, nil
}

func (c *Command) extractUniqueChannels(schema messageflow.Schema) []string {
	channelMap := make(map[string]bool)

	for _, service := range schema.Services {
		for _, operation := range service.Operation {
			channelMap[operation.Channel.Name] = true
			if operation.Reply != nil {
				channelMap[operation.Reply.Name] = true
			}
		}
	}

	channels := make([]string, 0, len(channelMap))
	for channel := range channelMap {
		channels = append(channels, channel)
	}

	sort.Strings(channels)
	return channels
}

func (c *Command) createREADMEContent(schema messageflow.Schema, title string) (string, error) {
	tmpl, err := template.New("readme.tmpl").Funcs(template.FuncMap{
		"Anchor": func(name string) string {
			return sanitizeAnchor(name)
		},
	}).ParseFS(readmeTemplateFS, "templates/readme.tmpl")
	if err != nil {
		return "", fmt.Errorf("error parsing README template: %w", err)
	}

	channels := c.extractUniqueChannels(schema)

	sort.Slice(schema.Services, func(i, j int) bool {
		return schema.Services[i].Name < schema.Services[j].Name
	})

	data := struct {
		Title    string
		Services []messageflow.Service
		Channels []string
	}{
		Title:    title,
		Services: schema.Services,
		Channels: channels,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing README template: %w", err)
	}

	return buf.String(), nil
}

func (c *Command) writeDiagramFiles(outputDir string, contextDiagram []byte, serviceDiagrams, channelDiagrams map[string][]byte) error {
	diagramsDir := filepath.Join(outputDir, "diagrams")
	if err := os.MkdirAll(diagramsDir, 0755); err != nil {
		return fmt.Errorf("error creating diagrams directory: %w", err)
	}

	contextPath := filepath.Join(diagramsDir, "context.svg")
	if err := os.WriteFile(contextPath, contextDiagram, 0644); err != nil {
		return fmt.Errorf("error writing context diagram: %w", err)
	}

	for serviceName, diagram := range serviceDiagrams {
		serviceAnchor := sanitizeAnchor(serviceName)
		servicePath := filepath.Join(diagramsDir, fmt.Sprintf("service_%s.svg", serviceAnchor))
		if err := os.WriteFile(servicePath, diagram, 0644); err != nil {
			return fmt.Errorf("error writing service diagram for %s: %w", serviceName, err)
		}
	}

	for channelName, diagram := range channelDiagrams {
		channelAnchor := sanitizeAnchor(channelName)
		channelPath := filepath.Join(diagramsDir, fmt.Sprintf("channel_%s.svg", channelAnchor))
		if err := os.WriteFile(channelPath, diagram, 0644); err != nil {
			return fmt.Errorf("error writing channel diagram for %s: %w", channelName, err)
		}
	}

	return nil
}

func sanitizeAnchor(name string) string {
	anchor := strings.ToLower(name)
	anchor = strings.ReplaceAll(anchor, " ", "-")
	anchor = strings.ReplaceAll(anchor, ".", "")
	anchor = strings.ReplaceAll(anchor, "_", "")
	anchor = strings.ReplaceAll(anchor, "{", "")
	anchor = strings.ReplaceAll(anchor, "}", "")
	return anchor
}
