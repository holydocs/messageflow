package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/denchenko/messageflow"
	"github.com/denchenko/messageflow/source/asyncapi"
	"github.com/denchenko/messageflow/target/d2"
)

func main() {
	targetType := flag.String("target", "d2", "Target type (d2)")
	formatToFile := flag.String("format-to-file", "", "Output file for the formatted schema")
	renderToFile := flag.String("render-to-file", "", "Output file for the rendered diagram")
	asyncAPIFilesPath := flag.String("asyncapi-files", "", "Paths to asyncapi files separated by comma")
	channel := flag.String("channel", "", "Channel")
	service := flag.String("service", "", "Service")
	formatMode := flag.String("format-mode", "service_channels", "Format mode")

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help ||
		*targetType == "" ||
		(*formatToFile == "" && *renderToFile == "") {
		printUsage()
		os.Exit(1)
	}

	target, err := pickTarget(*targetType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	targetCaps := target.Capabilities()

	if !targetCaps.Format {
		fmt.Fprintf(os.Stderr, "Error: Target doesn't support formatting\n")
		os.Exit(1)
	}

	if *renderToFile != "" && !targetCaps.Render {
		fmt.Fprintf(os.Stderr, "Error: Target doesn't support render\n")
		os.Exit(1)
	}

	ctx := context.Background()

	filePaths := strings.Split(*asyncAPIFilesPath, ",")
	schemas := make([]messageflow.Schema, 0, len(filePaths))

	for _, filePath := range filePaths {
		s, err := asyncapi.NewSource(strings.TrimSpace(filePath))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Creating schema source from %s %v\n", filePath, err)
			os.Exit(1)
		}

		schema, err := s.ExtractSchema(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Extracting schema from %s %v\n", filePath, err)
			os.Exit(1)
		}

		schemas = append(schemas, schema)
	}

	schema := messageflow.MergeSchemas(schemas...)

	formatOpts := messageflow.FormatOptions{
		Mode:    messageflow.FormatMode(*formatMode),
		Service: *service,
		Channel: *channel,
	}

	fs, err := target.FormatSchema(ctx, schema, formatOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Formatting schema %v\n", err)
		os.Exit(1)
	}

	if *formatToFile != "" {
		err = os.WriteFile(*formatToFile, fs.Data, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error : Writing to file %v\n", err)
			os.Exit(1)
		}
	}

	if *renderToFile != "" {
		diagram, err := target.RenderSchema(ctx, fs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Rendering schema %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile(*renderToFile, diagram, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Writing to file %v\n", err)
			os.Exit(1)
		}
	}
}

func pickTarget(targetType string) (messageflow.Target, error) {
	switch targetType {
	case "d2":
		return d2.NewTarget()
	}
	return nil, errors.New("unknown target")
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: messageflow [options]\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExample:\n")
	fmt.Fprintf(os.Stderr, "  messageflow --target d2 --render-to-file schema.svg --asyncapi-file asyncapi.yaml\n")
}
