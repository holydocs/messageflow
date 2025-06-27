package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/denchenko/messageflow"
	"github.com/denchenko/messageflow/pkg/schema/source/asyncapi"
)

func Load(ctx context.Context, paths []string) (messageflow.Schema, error) {
	schemas := make([]messageflow.Schema, 0, len(paths))

	for _, filePath := range paths {
		trimmedPath := strings.TrimSpace(filePath)

		s, err := asyncapi.NewSource(trimmedPath)
		if err != nil {
			return messageflow.Schema{}, fmt.Errorf("error creating schema source from %s: %w", trimmedPath, err)
		}

		schema, err := s.ExtractSchema(ctx)
		if err != nil {
			return messageflow.Schema{}, fmt.Errorf("error extracting schema from %s: %w", trimmedPath, err)
		}

		schemas = append(schemas, schema)
	}

	return messageflow.MergeSchemas(schemas...), nil
}
