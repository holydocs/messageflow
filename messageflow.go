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

// Schema defines the structure of a message flow schema containing services and their operations.
type Schema struct {
	Services []Service `json:"services"`
}

// Service represents a service in the message flow with its name and operations.
type Service struct {
	Name      string      `json:"name"`
	Operation []Operation `json:"operations"`
}

// Action represents the type of operation that can be performed on a channel.
type Action string

const (
	ActionSend    Action = "send"
	ActionReceive Action = "receive"
)

// Channel represents a communication channel with a name and message type.
type Channel struct {
	Name    string `json:"name"`
	Message string `json:"message"`
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
	FormatSchema(ctx context.Context, s Schema) (FormattedSchema, error)
}

// SchemaRenderer interface defines the contract for rendering formatted schemas.
type SchemaRenderer interface {
	RenderSchema(ctx context.Context, fs FormattedSchema) ([]byte, error)
}

// UnsupportedFormatError represents an error when an unsupported format is provided.
type UnsupportedFormatError struct {
	given    TargetType
	expected TargetType
}

// NewUnsupportedFormatError creates a new UnsupportedFormatError.
func NewUnsupportedFormatError(given, expected TargetType) error {
	return &UnsupportedFormatError{
		given:    given,
		expected: expected,
	}
}

// Error implements the error interface for UnsupportedFormatError.
func (err *UnsupportedFormatError) Error() string {
	return fmt.Sprintf("%s format is not supported, %s expected", err.given, err.expected)
}
