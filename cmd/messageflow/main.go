package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/denchenko/messageflow"
	"github.com/denchenko/messageflow/source/asyncapi"
	"github.com/denchenko/messageflow/target/d2"
)

func main() {
	sourceType := flag.String("source", "asyncapi", "Source doc type")
	targetType := flag.String("target", "d2", "Target type (d2)")
	formatToFile := flag.String("format-to-file", "", "Output file for the formatted schema")
	renderToFile := flag.String("render-to-file", "", "Output file for the rendered diagram")
	asyncAPIFilePath := flag.String("asyncapi-file", "", "Path to asyncapi file")

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help ||
		*targetType == "" ||
		(*formatToFile == "" && *renderToFile == "") {
		printUsage()
		os.Exit(1)
	}

	source, err := pickSource(*sourceType, *asyncAPIFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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

	schema, err := source.ExtractSchema(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Extracting schema %v\n", err)
		os.Exit(1)
	}

	fs, err := target.FormatSchema(ctx, schema)
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

func pickSource(sourceType, filePath string) (messageflow.Source, error) {
	switch sourceType {
	case "asyncapi":
		return asyncapi.NewSource(filePath)
	}
	return nil, errors.New("unknown source")
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
