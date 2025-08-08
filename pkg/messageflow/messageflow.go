// Package messageflow provides tools for visualizing AsyncAPI specifications.
// It enables parsing AsyncAPI documents and transforming them into visual format
// to help understand message flows and service interactions.
package messageflow

import (
	"context"
	"fmt"
	"sort"
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
	FormatModeServiceServices = FormatMode("service_services")
)

type FormatOptions struct {
	Mode         FormatMode
	Service      string
	Channel      string
	OmitPayloads bool
}

// Schema defines the structure of a message flow schema containing services and their operations.
type Schema struct {
	Services []Service `json:"services"`
}

// Sort sorts the services and their operations in a consistent order.
func (s *Schema) Sort() {
	for i := range s.Services {
		sort.Slice(s.Services[i].Operation, func(j, k int) bool {
			op1 := s.Services[i].Operation[j]
			op2 := s.Services[i].Operation[k]

			if op1.Action != op2.Action {
				return op1.Action < op2.Action
			}

			if op1.Channel.Name != op2.Channel.Name {
				return op1.Channel.Name < op2.Channel.Name
			}

			if len(op1.Channel.Messages) > 0 && len(op2.Channel.Messages) > 0 {
				return op1.Channel.Messages[0].Name < op2.Channel.Messages[0].Name
			}

			return len(op1.Channel.Messages) > len(op2.Channel.Messages)
		})
	}

	sort.Slice(s.Services, func(i, j int) bool {
		return s.Services[i].Name < s.Services[j].Name
	})
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

// Channel represents a communication channel with a name and messages.
type Channel struct {
	Name     string    `json:"name"`
	Messages []Message `json:"messages"`
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
	Category  string     `json:"category"`
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
					key := operationKey(op)
					opMap[key] = op
				}

				for _, op := range service.Operation {
					key := operationKey(op)
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
				Details:   fmt.Sprintf("'%s' was added", newService.Name),
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
				Details:   fmt.Sprintf("'%s' was removed", name),
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
				Category: "channel",
				Name:     fmt.Sprintf("%s:%s", newService.Name, key),
				Details: fmt.Sprintf(
					"'%s' on channel '%s' was added to service '%s'",
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
				Category: "channel",
				Name:     fmt.Sprintf("%s:%s", oldService.Name, key),
				Details: fmt.Sprintf(
					"'%s' on channel '%s' was removed from service '%s'",
					oldOp.Action, oldOp.Channel.Name, oldService.Name,
				),
				Timestamp: timestamp,
			})
		} else {
			newOp := newOps[key]
			// Compare channel messages
			if !cmp.Equal(oldOp.Channel.Messages, newOp.Channel.Messages) {
				diff := cmp.Diff(
					oldOp.Channel.Messages,
					newOp.Channel.Messages,
				)

				changes = append(changes, Change{
					Type:     ChangeTypeChanged,
					Category: "message",
					Name:     fmt.Sprintf("%s:%s", newService.Name, key),
					Details: fmt.Sprintf(
						"Messages changed for operation '%s' on channel '%s' in service '%s'",
						newOp.Action, newOp.Channel.Name, newService.Name,
					),
					Diff:      diff,
					Timestamp: timestamp,
				})
			}

			if oldOp.Reply != nil && newOp.Reply != nil {
				if !cmp.Equal(oldOp.Reply.Messages, newOp.Reply.Messages) {
					diff := cmp.Diff(
						oldOp.Reply.Messages,
						newOp.Reply.Messages,
					)

					changes = append(changes, Change{
						Type:     ChangeTypeChanged,
						Category: "message",
						Name:     fmt.Sprintf("%s:%s:reply", newService.Name, key),
						Details: fmt.Sprintf(
							"Reply messages changed for operation '%s' on channel '%s' in service '%s'",
							newOp.Action, newOp.Channel.Name, newService.Name,
						),
						Diff:      diff,
						Timestamp: timestamp,
					})
				}
			} else if oldOp.Reply != nil && newOp.Reply == nil {
				changes = append(changes, Change{
					Type:     ChangeTypeRemoved,
					Category: "channel",
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
					Category: "channel",
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
	messageName := ""
	if len(op.Channel.Messages) > 0 {
		messageName = op.Channel.Messages[0].Name
	}

	key := fmt.Sprintf("%s-%s-%s", op.Action, op.Channel.Name, messageName)
	if op.Reply != nil {
		replyMessageName := ""
		if len(op.Reply.Messages) > 0 {
			replyMessageName = op.Reply.Messages[0].Name
		}
		key += fmt.Sprintf("-reply-%s-%s", op.Reply.Name, replyMessageName)
	}
	return key
}
