// Package messageflow provides tools for visualizing AsyncAPI specifications.
// It enables parsing AsyncAPI documents and transforming them into visual format
// to help understand message flows and service interactions.
package messageflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
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

// ChangeType represents the type of change that occurred.
type ChangeType string

const (
	ChangeTypeAdded   ChangeType = "added"
	ChangeTypeRemoved ChangeType = "removed"
	ChangeTypeChanged ChangeType = "changed"
)

// Change represents a single change in the schema.
type Change struct {
	Type      ChangeType `json:"type"`
	Category  string     `json:"category"` // "service", "channel", "message"
	Name      string     `json:"name"`
	Details   string     `json:"details,omitempty"`
	Diff      string     `json:"diff,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// Changelog represents a collection of changes with a version and date.
type Changelog struct {
	Date    time.Time `json:"date"`
	Changes []Change  `json:"changes"`
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

// CompareSchemas compares two schemas and returns a changelog of differences.
func CompareSchemas(oldSchema, newSchema Schema) Changelog {
	changes := []Change{}
	now := time.Now()

	oldServices := make(map[string]Service)
	newServices := make(map[string]Service)

	for _, service := range oldSchema.Services {
		oldServices[service.Name] = service
	}

	for _, service := range newSchema.Services {
		newServices[service.Name] = service
	}

	for name, newService := range newServices {
		if _, exists := oldServices[name]; !exists {
			changes = append(changes, Change{
				Type:      ChangeTypeAdded,
				Category:  "service",
				Name:      name,
				Details:   fmt.Sprintf("Service '%s' was added", newService.Name),
				Timestamp: now,
			})
		}
	}

	for name, oldService := range oldServices {
		if _, exists := newServices[name]; !exists {
			changes = append(changes, Change{
				Type:      ChangeTypeRemoved,
				Category:  "service",
				Name:      name,
				Details:   fmt.Sprintf("Service '%s' was removed", name),
				Timestamp: now,
			})
		} else {
			// Compare operations within the same service
			serviceChanges := compareServiceOperations(oldService, newServices[name], now)
			changes = append(changes, serviceChanges...)
		}
	}

	return Changelog{
		Date:    now,
		Changes: changes,
	}
}

func compareServiceOperations(oldService, newService Service, timestamp time.Time) []Change {
	changes := []Change{}

	oldOps := make(map[string]Operation)
	newOps := make(map[string]Operation)

	for _, op := range oldService.Operation {
		key := operationKey(op)
		oldOps[key] = op
	}

	for _, op := range newService.Operation {
		key := operationKey(op)
		newOps[key] = op
	}

	for key, newOp := range newOps {
		if _, exists := oldOps[key]; !exists {
			changes = append(changes, Change{
				Type:     ChangeTypeAdded,
				Category: "operation",
				Name:     fmt.Sprintf("%s:%s", newService.Name, key),
				Details: fmt.Sprintf(
					"Operation '%s' on channel '%s' was added to service '%s'",
					newOp.Action, newOp.Channel.Name, newService.Name,
				),
				Timestamp: timestamp,
			})
		}
	}

	for key, oldOp := range oldOps {
		if _, exists := newOps[key]; !exists {
			changes = append(changes, Change{
				Type:     ChangeTypeRemoved,
				Category: "operation",
				Name:     fmt.Sprintf("%s:%s", oldService.Name, key),
				Details: fmt.Sprintf(
					"Operation '%s' on channel '%s' was removed from service '%s'",
					oldOp.Action, oldOp.Channel.Name, oldService.Name,
				),
				Timestamp: timestamp,
			})
		} else {
			newOp := newOps[key]
			if !cmp.Equal(oldOp.Channel.Message.Payload, newOp.Channel.Message.Payload) {
				diff := cmp.Diff(
					oldOp.Channel.Message.Payload,
					newOp.Channel.Message.Payload,
				)

				changes = append(changes, Change{
					Type:     ChangeTypeChanged,
					Category: "message",
					Name:     fmt.Sprintf("%s:%s", newService.Name, key),
					Details: fmt.Sprintf(
						"Message payload changed for operation '%s' on channel '%s' in service '%s'",
						newOp.Action, newOp.Channel.Name, newService.Name,
					),
					Diff:      diff,
					Timestamp: timestamp,
				})
			}

			if oldOp.Reply != nil && newOp.Reply != nil {
				if !cmp.Equal(oldOp.Reply.Message.Payload, newOp.Reply.Message.Payload) {
					diff := cmp.Diff(
						oldOp.Reply.Message.Payload,
						newOp.Reply.Message.Payload,
					)

					changes = append(changes, Change{
						Type:     ChangeTypeChanged,
						Category: "message",
						Name:     fmt.Sprintf("%s:%s:reply", newService.Name, key),
						Details: fmt.Sprintf(
							"Reply message payload changed for operation '%s' on channel '%s' in service '%s'",
							newOp.Action, newOp.Channel.Name, newService.Name,
						),
						Diff:      diff,
						Timestamp: timestamp,
					})
				}
			} else if oldOp.Reply != nil && newOp.Reply == nil {
				changes = append(changes, Change{
					Type:     ChangeTypeRemoved,
					Category: "operation",
					Name:     fmt.Sprintf("%s:%s:reply", newService.Name, key),
					Details: fmt.Sprintf(
						"Reply channel removed for operation '%s' on channel '%s' in service '%s'",
						newOp.Action, newOp.Channel.Name, newService.Name,
					),
					Timestamp: timestamp,
				})
			} else if oldOp.Reply == nil && newOp.Reply != nil {
				changes = append(changes, Change{
					Type:     ChangeTypeAdded,
					Category: "operation",
					Name:     fmt.Sprintf("%s:%s:reply", newService.Name, key),
					Details: fmt.Sprintf(
						"Reply channel added for operation '%s' on channel '%s' in service '%s'",
						newOp.Action, newOp.Channel.Name, newService.Name,
					),
					Timestamp: timestamp,
				})
			}
		}
	}

	return changes
}

func operationKey(op Operation) string {
	key := fmt.Sprintf("%s-%s-%s", op.Action, op.Channel.Name, op.Channel.Message.Name)
	if op.Reply != nil {
		key += fmt.Sprintf("-reply-%s-%s", op.Reply.Name, op.Reply.Message.Name)
	}
	return key
}
