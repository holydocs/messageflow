// Package d2 provides functionality for generating and rendering D2 diagrams.
package d2

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"text/template"

	"github.com/denchenko/messageflow"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

// targetType defines the schema format type for D2 diagrams
const targetType = messageflow.TargetType("d2")

var (
	//go:embed templates/service_channels.tmpl
	serviceChannelsTemplateFS embed.FS

	//go:embed templates/channel_services.tmpl
	channelServicesTemplateFS embed.FS
)

// Ensure Target implements messageflow interfaces.
var (
	_ messageflow.Target = (*Target)(nil)
)

// Target handles the generation and rendering of D2 diagrams from message flow schemas.
type Target struct {
	serviceChannelsTemplate *template.Template
	channelServicesTemplate *template.Template
	renderOpts              *d2svg.RenderOpts
	compileOpts             *d2lib.CompileOptions
}

// TargetOpt is a function type that allows customization of a Target instance.
type TargetOpt func(*Target)

// WithRenderOpts returns a TargetOpt that sets the rendering options for the D2 diagram.
// These options control aspects such as padding, theme, and other visual properties.
func WithRenderOpts(renderOpts *d2svg.RenderOpts) TargetOpt {
	return func(t *Target) {
		t.renderOpts = renderOpts
	}
}

// WithCompileOpts returns a TargetOpt that sets the compilation options for the D2 diagram.
// These options control the layout and measurement aspects of the diagram generation.
func WithCompileOpts(compileOpts *d2lib.CompileOptions) TargetOpt {
	return func(t *Target) {
		t.compileOpts = compileOpts
	}
}

// NewTarget creates a new D2 diagram formatter instance.
// It initializes the template from the embedded schema.tmpl file and sets up default
// rendering and compilation options. The formatter uses the ELK layout engine for
// diagram arrangement.
func NewTarget() (*Target, error) {
	serviceChannelsTemplate, err := template.ParseFS(serviceChannelsTemplateFS, "templates/service_channels.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	channelServicesTemplate, err := template.ParseFS(channelServicesTemplateFS, "templates/channel_services.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("creating ruler: %w", err)
	}

	layoutResolver := func(_ string) (d2graph.LayoutGraph, error) {
		return d2elklayout.DefaultLayout, nil
	}

	return &Target{
		serviceChannelsTemplate: serviceChannelsTemplate,
		channelServicesTemplate: channelServicesTemplate,
		renderOpts: &d2svg.RenderOpts{
			Pad:     go2.Pointer(int64(5)),
			ThemeID: &d2themescatalog.Terminal.ID,
		},
		compileOpts: &d2lib.CompileOptions{
			LayoutResolver: layoutResolver,
			Ruler:          ruler,
		},
	}, nil
}

// Capabilities returns target capabilities.
func (t *Target) Capabilities() messageflow.TargetCapabilities {
	return messageflow.TargetCapabilities{
		Format: true,
		Render: true,
	}
}

type channelServicesPayload struct {
	Channel      string
	Message      string
	ReplyMessage *string
	Senders      []string
	Receivers    []string
}

func (t *Target) FormatSchema(
	_ context.Context,
	s messageflow.Schema,
	opts messageflow.FormatOptions,
) (messageflow.FormattedSchema, error) {
	fs := messageflow.FormattedSchema{
		Type: targetType,
	}

	var buf bytes.Buffer

	switch opts.Mode {
	case messageflow.FormatModeServiceChannels:
		payload := prepareServiceChannelsPayload(s, opts.Service)

		err := t.serviceChannelsTemplate.Execute(&buf, payload)
		if err != nil {
			return messageflow.FormattedSchema{}, fmt.Errorf("executing service channels template: %w", err)
		}
	case messageflow.FormatModeChannelServices:
		payload := prepareChannelServicesPayload(s, opts.Channel)

		err := t.channelServicesTemplate.Execute(&buf, payload)
		if err != nil {
			return messageflow.FormattedSchema{}, fmt.Errorf("executing channel services template: %w", err)
		}
	default:
		return messageflow.FormattedSchema{}, messageflow.NewUnsupportedFormatModeError(opts.Mode, []messageflow.FormatMode{
			messageflow.FormatModeServiceChannels,
			messageflow.FormatModeChannelServices,
		})
	}

	fs.Data = buf.Bytes()

	return fs, nil
}

// RenderSchema renders a formatted D2 diagram to SVG format.
func (t *Target) RenderSchema(ctx context.Context, s messageflow.FormattedSchema) ([]byte, error) {
	if s.Type != targetType {
		return nil, messageflow.NewUnsupportedFormatError(s.Type, targetType)
	}

	ctx = log.WithDefault(ctx)

	diagram, _, err := d2lib.Compile(ctx, string(s.Data), t.compileOpts, t.renderOpts)
	if err != nil {
		return nil, fmt.Errorf("compiling diagram: %w", err)
	}

	out, err := d2svg.Render(diagram, t.renderOpts)
	if err != nil {
		return nil, fmt.Errorf("rendering diagram: %w", err)
	}

	return out, nil
}

func prepareServiceChannelsPayload(s messageflow.Schema, serviceName string) messageflow.Service {
	if serviceName == "" && len(s.Services) == 1 {
		return s.Services[0]
	}

	for _, service := range s.Services {
		if service.Name == serviceName {
			return service
		}
	}

	return messageflow.Service{}
}

func prepareChannelServicesPayload(s messageflow.Schema, channel string) channelServicesPayload {
	payload := channelServicesPayload{
		Channel: channel,
	}

	for _, service := range s.Services {
		for _, op := range service.Operation {
			if op.Channel.Name == channel {
				switch op.Action {
				case messageflow.ActionSend:
					payload.Senders = append(payload.Senders, service.Name)
				case messageflow.ActionReceive:
					payload.Receivers = append(payload.Receivers, service.Name)
				}

				if len(payload.Message) < len(op.Channel.Message) {
					payload.Message = op.Channel.Message
				}

				if op.Reply != nil && (payload.ReplyMessage == nil ||
					(len(*payload.ReplyMessage) < len(op.Reply.Message))) {
					payload.ReplyMessage = &op.Reply.Message
				}
			}
		}
	}

	return payload
}
