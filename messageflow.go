// Package messageflow provides tools for visualizing AsyncAPI specifications.
// It enables parsing AsyncAPI documents and transforming them into visual format
// to help understand message flows and service interactions.
package messageflow

import (
	"context"
	"fmt"
)

// TargetType represents the type of target format for schema conversion.
type TargetType string

// FormatMode represents the mode of format for schema.
type FormatMode string

const (
	FormatModeContextServices = FormatMode("context_services")
	FormatModeServiceChannels = FormatMode("service_channels")
	FormatModeChannelServices = FormatMode("channel_services")
)

type FormatOptions struct {
	Mode    FormatMode
	Service string
	Channel string
}

// Schema defines the structure of a message flow schema containing services and their operations.
type Schema struct {
	Services []Service `json:"services"`
}

// Service represents a service in the message flow with its name and operations.
type Service struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Operation   []Operation `json:"operations"`
}

// Action represents the type of operation that can be performed on a channel.
type Action string

const (
	ActionSend    Action = "send"
	ActionReceive Action = "receive"
)

// Message represents a message with a name and payload.
type Message struct {
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

// Channel represents a communication channel with a name and message.
type Channel struct {
	Name    string  `json:"name"`
	Message Message `json:"message"`
}

// Operation defines an action to be performed on a channel, optionally with a reply channel.
type Operation struct {
	Action  Action   `json:"action"`
	Channel Channel  `json:"channel"`
	Reply   *Channel `json:"reply,omitempty"`
}

// FormattedSchema represents a schema that has been formatted for a specific target type.
type FormattedSchema struct {
	Type TargetType `json:"type"`
	Data []byte     `json:"data"`
}

// TargetCapabilities represents the capabilities of a Target implementation.
type TargetCapabilities struct {
	Format bool
	Render bool
}

// Source interface defines the contract for schema extraction.
type Source interface {
	SchemaExtractor
}

// Target interface defines the contract for schema formatting and rendering.
type Target interface {
	SchemaFormatter
	SchemaRenderer
	Capabilities() TargetCapabilities
}

// SchemaExtractor interface defines the contract for extracting schemas.
type SchemaExtractor interface {
	ExtractSchema(ctx context.Context) (Schema, error)
}

// SchemaFormatter interface defines the contract for formatting schemas.
type SchemaFormatter interface {
	FormatSchema(ctx context.Context, s Schema, opts FormatOptions) (FormattedSchema, error)
}

// SchemaRenderer interface defines the contract for rendering formatted schemas.
type SchemaRenderer interface {
	RenderSchema(ctx context.Context, fs FormattedSchema) ([]byte, error)
}

// MergeSchemas combines multiple Schema objects into a single Schema.
func MergeSchemas(schemas ...Schema) Schema {
	if len(schemas) == 0 {
		return Schema{Services: []Service{}}
	}

	serviceMap := make(map[string]Service)

	for _, schema := range schemas {
		for _, service := range schema.Services {
			if existingService, exists := serviceMap[service.Name]; exists {
				opMap := make(map[string]Operation)

				for _, op := range existingService.Operation {
					key := fmt.Sprintf("%s-%s-%s", op.Action, op.Channel.Name, op.Channel.Message.Name)
					opMap[key] = op
				}

				for _, op := range service.Operation {
					key := fmt.Sprintf("%s-%s-%s", op.Action, op.Channel.Name, op.Channel.Message.Name)
					opMap[key] = op
				}

				mergedOps := make([]Operation, 0, len(opMap))
				for _, op := range opMap {
					mergedOps = append(mergedOps, op)
				}

				existingService.Operation = mergedOps
				serviceMap[service.Name] = existingService
			} else {
				serviceMap[service.Name] = service
			}
		}
	}

	mergedServices := make([]Service, 0, len(serviceMap))
	for _, service := range serviceMap {
		mergedServices = append(mergedServices, service)
	}

	return Schema{Services: mergedServices}
}
