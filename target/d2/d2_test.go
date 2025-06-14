package d2

import (
	"context"
	_ "embed"
	"testing"

	"github.com/denchenko/messageflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	//go:embed testdata/schema.d2
	testSchema []byte

	//go:embed testdata/schema.svg
	testSVG []byte
)

func TestFormatSchema(t *testing.T) {
	t.Parallel()

	schema := messageflow.Schema{
		Services: []messageflow.Service{
			{
				Name: "Notification Service",
				Operation: []messageflow.Operation{
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "notification.preferences.get",
							Message: `{
  "user_id": "string[uuid]"
}`,
						},
						Reply: &messageflow.Channel{
							Name: "notification.preferences.get",
							Message: `{
  "preferences": {
    "categories": {
      "marketing": "boolean",
      "security": "boolean",
      "updates": "boolean"
    },
    "email_enabled": "boolean",
    "push_enabled": "boolean",
    "quiet_hours": {
      "enabled": "boolean",
      "end": "string[time]",
      "start": "string[time]"
    },
    "sms_enabled": "boolean"
  },
  "updated_at": "string[date-time]"
}`,
						},
					},
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "notification.preferences.update",
							Message: `{
  "preferences": {
    "categories": {
      "marketing": "boolean",
      "security": "boolean",
      "updates": "boolean"
    },
    "email_enabled": "boolean",
    "push_enabled": "boolean",
    "quiet_hours": {
      "enabled": "boolean",
      "end": "string[time]",
      "start": "string[time]"
    },
    "sms_enabled": "boolean"
  },
  "updated_at": "string[date-time]",
  "user_id": "string[uuid]"
}`,
						},
					},
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "notification.user.{user_id}.push",
							Message: `{
  "body": "string",
  "created_at": "string[date-time]",
  "data": "object",
  "notification_id": "string[uuid]",
  "priority": "string[enum:low,normal,high]",
  "title": "string",
  "user_id": "string[uuid]"
}`,
						},
					},
					{
						Action: messageflow.ActionSend,
						Channel: messageflow.Channel{
							Name: "user.info.request",
							Message: `{
  "user_id": "string[uuid]"
}`,
						},
						Reply: &messageflow.Channel{
							Name: "user.info.request",
							Message: `{
  "email": "string[email]",
  "error": {
    "code": "string",
    "message": "string"
  },
  "language": "string",
  "name": "string",
  "timezone": "string",
  "user_id": "string[uuid]"
}`,
						},
					},
					{
						Action: messageflow.ActionSend,
						Channel: messageflow.Channel{
							Name: "notification.analytics",
							Message: `{
  "event_id": "string[uuid]",
  "event_type": "string[enum:notification_sent,notification_opened,notification_clicked]",
  "metadata": {
    "environment": "string[enum:development,staging,production]",
    "platform": "string[enum:ios,android,web]",
    "source": "string[enum:mobile,web,api]",
    "version": "string"
  },
  "notification_id": "string[uuid]",
  "timestamp": "string[date-time]",
  "user_id": "string[uuid]"
}`,
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	target, err := NewTarget()
	require.NoError(t, err)

	actual, err := target.FormatSchema(ctx, schema)
	require.NoError(t, err)

	expected := messageflow.FormattedSchema{
		Type: "d2",
		Data: testSchema,
	}
	assert.Equal(t, expected, actual)
}

func TestRenderSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	target, err := NewTarget()
	require.NoError(t, err)

	actual, err := target.RenderSchema(ctx, messageflow.FormattedSchema{
		Type: "d2",
		Data: testSchema,
	})
	require.NoError(t, err)

	assert.Equal(t, testSVG, actual)
}
