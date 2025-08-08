// Package asyncapi provides functionality for extracting message flow schemas from AsyncAPI specifications.
package asyncapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/holydocs/messageflow/pkg/messageflow"
	"github.com/lerenn/asyncapi-codegen/pkg/asyncapi/parser"
	asyncapiv3 "github.com/lerenn/asyncapi-codegen/pkg/asyncapi/v3"
)

// Ensure Source implements messageflow interfaces.
var (
	_ messageflow.Source = (*Source)(nil)
)

// Source represents a AsyncAPI source for schema extraction.
type Source struct {
	path string
}

// NewSource creates a new AsyncAPI source from a multiple paths to specifications.
func NewSource(path string) (*Source, error) {
	return &Source{
		path: path,
	}, nil
}

// ExtractSchema extracts messageflow schema from AsyncAPI specifications.
func (s *Source) ExtractSchema(_ context.Context) (messageflow.Schema, error) {
	spec, err := s.loadAndProcessSpec()
	if err != nil {
		return messageflow.Schema{}, err
	}

	service := s.createServiceFromSpec(spec)

	return messageflow.Schema{
		Services: []messageflow.Service{service},
	}, nil
}

// loadAndProcessSpec loads and processes the AsyncAPI specification from file.
func (s *Source) loadAndProcessSpec() (*asyncapiv3.Specification, error) {
	spec, err := parser.FromFile(parser.FromFileParams{
		Path: s.path,
	})
	if err != nil {
		return nil, fmt.Errorf("parsing AsyncAPI spec from %s: %w", s.path, err)
	}

	if err := spec.Process(); err != nil {
		return nil, fmt.Errorf("processing AsyncAPI spec from %s: %w", s.path, err)
	}

	v3Spec, err := asyncapiv3.FromUnknownVersion(spec)
	if err != nil {
		return nil, fmt.Errorf("converting to v3 spec from %s: %w", s.path, err)
	}

	return v3Spec, nil
}

// createServiceFromSpec creates a messageflow.Service from an AsyncAPI v3 specification.
func (s *Source) createServiceFromSpec(spec *asyncapiv3.Specification) messageflow.Service {
	service := messageflow.Service{
		Name:        spec.Info.Title,
		Description: spec.Info.Description,
		Operation:   make([]messageflow.Operation, 0),
	}

	for _, op := range spec.Operations {
		operation := s.createOperation(op)
		if operation != nil {
			service.Operation = append(service.Operation, *operation)
		}
	}

	return service
}

// createOperation creates a messageflow.Operation from an AsyncAPI operation.
func (s *Source) createOperation(op *asyncapiv3.Operation) *messageflow.Operation {
	channel := op.Channel.Follow()
	if channel == nil {
		return nil
	}

	mainMessages := s.extractMainMessages(op)
	if len(mainMessages) == 0 {
		return nil
	}

	operation := messageflow.Operation{
		Action: messageflow.Action(op.Action),
		Channel: messageflow.Channel{
			Name:     channel.Address,
			Messages: mainMessages,
		},
	}

	if op.Reply != nil {
		replyMessages := s.extractReplyMessages(op)
		if len(replyMessages) > 0 {
			replyChannel := op.Reply.Channel.Follow()
			if replyChannel != nil {
				operation.Reply = &messageflow.Channel{
					Name:     replyChannel.Address,
					Messages: replyMessages,
				}
			}
		}
	}

	return &operation
}

// extractMainMessages extracts all main messages from an operation.
func (s *Source) extractMainMessages(op *asyncapiv3.Operation) []messageflow.Message {
	messages := make([]messageflow.Message, 0)

	for _, msgRef := range op.Messages {
		if msgRef == nil {
			continue
		}

		msg := msgRef
		for msg != nil && msg.Payload == nil && msg.ReferenceTo != nil {
			msg = msg.ReferenceTo
		}

		if msg == nil || msg.Payload == nil {
			continue
		}

		jsonSchema, err := jsonMessage(msg.Payload)
		if err != nil {
			continue
		}

		messageName := s.extractMessageName(msg)
		messages = append(messages, messageflow.Message{
			Name:    messageName,
			Payload: jsonSchema,
		})
	}

	return messages
}

// extractReplyMessages extracts all reply messages from an operation.
func (s *Source) extractReplyMessages(op *asyncapiv3.Operation) []messageflow.Message {
	if op.Reply == nil {
		return nil
	}

	messages := make([]messageflow.Message, 0)

	for _, msgRef := range op.Reply.Messages {
		if msgRef == nil {
			continue
		}

		msg := msgRef
		for msg != nil && msg.Payload == nil && msg.ReferenceTo != nil {
			msg = msg.ReferenceTo
		}

		if msg == nil || msg.Payload == nil {
			continue
		}

		jsonSchema, err := jsonMessage(msg.Payload)
		if err != nil {
			continue
		}

		messageName := s.extractMessageName(msg)
		messages = append(messages, messageflow.Message{
			Name:    messageName,
			Payload: jsonSchema,
		})
	}

	return messages
}

// extractMessageName extracts the message name from a message reference.
func (s *Source) extractMessageName(msg *asyncapiv3.Message) string {
	if msg == nil {
		return ""
	}

	// Try to get name from the message itself
	if msg.Name != "" {
		return msg.Name
	}

	// If no name in the message, try to get it from the title
	if msg.Title != "" {
		return msg.Title
	}

	// If still no name, try to get it from the summary
	if msg.Summary != "" {
		return msg.Summary
	}

	// Fallback to a generic name
	return "UnknownMessage"
}

// jsonMessage converts an AsyncAPI schema into a pretty-printed JSON string.
func jsonMessage(schema *asyncapiv3.Schema) (string, error) {
	if schema == nil {
		return "", nil
	}

	for schema.ReferenceTo != nil {
		schema = schema.ReferenceTo
	}

	schemaMap := make(map[string]any)

	if len(schema.Properties) > 0 {
		props := make(map[string]any)
		for name, prop := range schema.Properties {
			for prop.ReferenceTo != nil {
				prop = prop.ReferenceTo
			}
			props[name] = getTypeString(prop)
		}
		schemaMap = props
	}

	data, err := json.MarshalIndent(schemaMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling schema: %w", err)
	}

	return string(data), nil
}

// getTypeString returns a string representation of the schema type
func getTypeString(schema *asyncapiv3.Schema) any {
	if schema == nil {
		return "string"
	}

	if schema.ReferenceTo != nil {
		schema = schema.ReferenceTo
	}

	if schema.Type == "array" {
		if schema.Items == nil {
			return []any{}
		}
		if schema.Items.ReferenceTo != nil {
			schema.Items = schema.Items.ReferenceTo
		}
		return []any{getTypeString(schema.Items)}
	}

	if schema.Type == "object" {
		if len(schema.Properties) == 0 {
			return "object"
		}
		props := make(map[string]any)
		for name, prop := range schema.Properties {
			props[name] = getTypeString(prop)
		}
		return props
	}

	if schema.Type != "" {
		if schema.Format != "" {
			return schema.Type + "[" + schema.Format + "]"
		}
		if len(schema.Enum) > 0 {
			enumValues := make([]string, len(schema.Enum))
			for i, v := range schema.Enum {
				enumValues[i] = fmt.Sprintf("%v", v)
			}
			return schema.Type + "[enum:" + strings.Join(enumValues, ",") + "]"
		}
		return schema.Type
	}

	return "string"
}
