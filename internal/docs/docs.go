package docs

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/denchenko/messageflow"
	"golang.org/x/sync/errgroup"
)

//go:embed templates/readme.tmpl
var readmeTemplateFS embed.FS

type Metadata struct {
	Schema     messageflow.Schema      `json:"schema"`
	Changelogs []messageflow.Changelog `json:"changelogs"`
}

func Generate(
	ctx context.Context,
	schema messageflow.Schema,
	target messageflow.Target,
	title, outputDir string,
) error {
	metadata, err := processMetadata(schema, outputDir)
	if err != nil {
		return fmt.Errorf("error processing metadata: %w", err)
	}

	if err := generateDiagrams(ctx, schema, target, outputDir); err != nil {
		return fmt.Errorf("error generating diagrams: %w", err)
	}

	if err := createREADMEContent(schema, title, metadata.Changelogs, outputDir); err != nil {
		return fmt.Errorf("error creating README content: %w", err)
	}

	return nil
}

func processMetadata(schema messageflow.Schema, outputDir string) (*Metadata, error) {
	existingMetadata, err := readMetadata(outputDir)
	if err != nil {
		return nil, fmt.Errorf("error reading existing messageflow data: %w", err)
	}

	var newChangelog *messageflow.Changelog
	var existingChangelogs []messageflow.Changelog

	if existingMetadata != nil {
		changelog := messageflow.CompareSchemas(existingMetadata.Schema, schema)
		if len(changelog.Changes) > 0 {
			newChangelog = &changelog
		}
		existingChangelogs = existingMetadata.Changelogs
	}

	metadata := Metadata{
		Schema:     schema,
		Changelogs: existingChangelogs,
	}

	if newChangelog != nil {
		metadata.Changelogs = append(metadata.Changelogs, *newChangelog)
	}

	if err := writeMetadata(outputDir, metadata); err != nil {
		return nil, fmt.Errorf("error writing messageflow data: %w", err)
	}

	return &metadata, nil
}

func generateDiagrams(
	ctx context.Context,
	schema messageflow.Schema,
	target messageflow.Target,
	outputDir string,
) error {
	diagramsDir := filepath.Join(outputDir, "diagrams")
	if err := os.RemoveAll(diagramsDir); err != nil {
		return fmt.Errorf("error removing old diagrams directory: %w", err)
	}

	if err := os.MkdirAll(diagramsDir, 0755); err != nil {
		return fmt.Errorf("error creating diagrams directory: %w", err)
	}

	channels := extractUniqueChannels(schema)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return generateContextDiagram(ctx, schema, target, outputDir)
	})

	for _, service := range schema.Services {
		g.Go(func() error {
			return generateServiceServicesDiagram(ctx, schema, target, service.Name, outputDir)
		})
	}

	for _, channel := range channels {
		g.Go(func() error {
			return generateChannelServicesDiagram(ctx, schema, target, channel, outputDir)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error generating diagrams: %w", err)
	}

	return nil
}

func generateContextDiagram(
	ctx context.Context,
	schema messageflow.Schema,
	target messageflow.Target,
	outputDir string,
) error {
	formatOpts := messageflow.FormatOptions{
		Mode: messageflow.FormatModeContextServices,
	}

	formattedSchema, err := target.FormatSchema(ctx, schema, formatOpts)
	if err != nil {
		return fmt.Errorf("error formatting context schema: %w", err)
	}

	diagram, err := target.RenderSchema(ctx, formattedSchema)
	if err != nil {
		return fmt.Errorf("error rendering context diagram: %w", err)
	}

	contextPath := filepath.Join(outputDir, "diagrams", "context.svg")
	if err := os.WriteFile(contextPath, diagram, 0644); err != nil {
		return fmt.Errorf("error writing context diagram: %w", err)
	}

	return nil
}

func generateServiceServicesDiagram(
	ctx context.Context,
	schema messageflow.Schema,
	target messageflow.Target,
	serviceName string,
	outputDir string,
) error {
	formatOpts := messageflow.FormatOptions{
		Mode:    messageflow.FormatModeServiceServices,
		Service: serviceName,
	}

	formattedSchema, err := target.FormatSchema(ctx, schema, formatOpts)
	if err != nil {
		return fmt.Errorf("error formatting service services schema: %w", err)
	}

	diagram, err := target.RenderSchema(ctx, formattedSchema)
	if err != nil {
		return fmt.Errorf("error rendering service services diagram: %w", err)
	}

	serviceAnchor := sanitizeAnchor(serviceName)
	servicePath := filepath.Join(outputDir, "diagrams", fmt.Sprintf("service_%s.svg", serviceAnchor))
	if err := os.WriteFile(servicePath, diagram, 0644); err != nil {
		return fmt.Errorf("error writing service diagram for %s: %w", serviceName, err)
	}

	return nil
}

func generateChannelServicesDiagram(
	ctx context.Context,
	schema messageflow.Schema,
	target messageflow.Target,
	channel string,
	outputDir string,
) error {
	formatOpts := messageflow.FormatOptions{
		Mode:         messageflow.FormatModeChannelServices,
		Channel:      channel,
		OmitPayloads: true,
	}

	formattedSchema, err := target.FormatSchema(ctx, schema, formatOpts)
	if err != nil {
		return fmt.Errorf("error formatting channel services schema: %w", err)
	}

	diagram, err := target.RenderSchema(ctx, formattedSchema)
	if err != nil {
		return fmt.Errorf("error rendering channel services diagram: %w", err)
	}

	channelAnchor := sanitizeAnchor(channel)
	channelPath := filepath.Join(outputDir, "diagrams", fmt.Sprintf("channel_%s.svg", channelAnchor))
	if err := os.WriteFile(channelPath, diagram, 0644); err != nil {
		return fmt.Errorf("error writing channel diagram for %s: %w", channel, err)
	}

	return nil
}

func extractUniqueChannels(schema messageflow.Schema) []string {
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

func createREADMEContent(schema messageflow.Schema, title string, changelogs []messageflow.Changelog, outputDir string) error {
	tmpl, err := template.New("readme.tmpl").Funcs(template.FuncMap{
		"Anchor": func(name string) string {
			return sanitizeAnchor(name)
		},
		"SortChangelogs": func(changelogs []messageflow.Changelog) []messageflow.Changelog {
			sorted := make([]messageflow.Changelog, len(changelogs))
			copy(sorted, changelogs)

			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Date.After(sorted[j].Date)
			})

			for i := range sorted {
				sort.Slice(sorted[i].Changes, func(a, b int) bool {
					if sorted[i].Changes[a].Type != sorted[i].Changes[b].Type {
						return string(sorted[i].Changes[a].Type) < string(sorted[i].Changes[b].Type)
					}

					if sorted[i].Changes[a].Category != sorted[i].Changes[b].Category {
						return sorted[i].Changes[a].Category < sorted[i].Changes[b].Category
					}

					return sorted[i].Changes[a].Name < sorted[i].Changes[b].Name
				})
			}

			return sorted
		},
	}).ParseFS(readmeTemplateFS, "templates/readme.tmpl")
	if err != nil {
		return fmt.Errorf("error parsing README template: %w", err)
	}

	channels := extractUniqueChannels(schema)
	channelInfo := extractChannelInfo(schema)

	sort.Slice(schema.Services, func(i, j int) bool {
		return schema.Services[i].Name < schema.Services[j].Name
	})

	data := struct {
		Title       string
		Services    []messageflow.Service
		Channels    []string
		ChannelInfo map[string]ChannelInfo
		Changelogs  []messageflow.Changelog
	}{
		Title:       title,
		Services:    schema.Services,
		Channels:    channels,
		ChannelInfo: channelInfo,
		Changelogs:  changelogs,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("error executing README template: %w", err)
	}

	readmePath := filepath.Join(outputDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("error writing README.md: %w", err)
	}

	return nil
}

// ChannelInfo represents information about a channel including its messages and payloads
type ChannelInfo struct {
	Messages []ChannelMessage
}

// ChannelMessage represents a message in a channel with its payload and direction
type ChannelMessage struct {
	Name      string
	Payload   string
	Direction string // "send" or "receive"
	Service   string
}

func extractChannelInfo(schema messageflow.Schema) map[string]ChannelInfo {
	channelInfo := make(map[string]ChannelInfo)

	channelOperations := make(map[string][]struct {
		service   string
		operation messageflow.Operation
	})

	for _, service := range schema.Services {
		for _, operation := range service.Operation {
			channelName := operation.Channel.Name
			channelOperations[channelName] = append(channelOperations[channelName], struct {
				service   string
				operation messageflow.Operation
			}{
				service:   service.Name,
				operation: operation,
			})
		}
	}

	// Select the most relevant message for each channel
	for channelName, operations := range channelOperations {
		info := ChannelInfo{
			Messages: []ChannelMessage{},
		}

		// Check if this is a req/reply pattern
		hasReply := false
		for _, op := range operations {
			if op.operation.Reply != nil {
				hasReply = true
				break
			}
		}

		if hasReply {
			// For req/reply pattern: include both request and reply messages
			for _, op := range operations {
				if op.operation.Reply != nil {
					info.Messages = append(info.Messages, ChannelMessage{
						Name:      op.operation.Channel.Message.Name,
						Payload:   op.operation.Channel.Message.Payload,
						Direction: "request",
						Service:   op.service,
					})
					info.Messages = append(info.Messages, ChannelMessage{
						Name:      op.operation.Reply.Message.Name,
						Payload:   op.operation.Reply.Message.Payload,
						Direction: "reply",
						Service:   op.service,
					})
					break
				}
			}
		} else {
			// For send/receive pattern: prefer the receive message
			receiveFound := false
			for _, op := range operations {
				if op.operation.Action == messageflow.ActionReceive {
					info.Messages = append(info.Messages, ChannelMessage{
						Name:      op.operation.Channel.Message.Name,
						Payload:   op.operation.Channel.Message.Payload,
						Direction: "receive",
						Service:   op.service,
					})
					receiveFound = true
					break
				}
			}

			// If no receive operation found, use the send message
			if !receiveFound {
				for _, op := range operations {
					if op.operation.Action == messageflow.ActionSend {
						info.Messages = append(info.Messages, ChannelMessage{
							Name:      op.operation.Channel.Message.Name,
							Payload:   op.operation.Channel.Message.Payload,
							Direction: "send",
							Service:   op.service,
						})
						break
					}
				}
			}
		}

		channelInfo[channelName] = info
	}

	return channelInfo
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

func readMetadata(outputDir string) (*Metadata, error) {
	dataPath := filepath.Join(outputDir, "messageflow.json")

	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return nil, nil
	}

	fileData, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("error reading messageflow data file: %w", err)
	}

	var messageFlowData Metadata
	if err := json.Unmarshal(fileData, &messageFlowData); err != nil {
		return nil, fmt.Errorf("error unmarshaling messageflow data: %w", err)
	}

	return &messageFlowData, nil
}

func writeMetadata(outputDir string, data Metadata) error {
	dataPath := filepath.Join(outputDir, "messageflow.json")

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling messageflow data: %w", err)
	}

	if err := os.WriteFile(dataPath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing messageflow data file: %w", err)
	}

	return nil
}
