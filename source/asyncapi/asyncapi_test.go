package asyncapi

import (
	"context"
	_ "embed"
	"sort"
	"testing"

	"github.com/denchenko/messageflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractSchema(t *testing.T) {
	ctx := context.Background()
	source, err := NewSource("testdata/notification.yaml")
	require.NoError(t, err)
	actual, err := source.ExtractSchema(ctx)
	require.NoError(t, err)

	expected := messageflow.Schema{
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

	sortSchema(&expected)
	sortSchema(&actual)

	assert.Equal(t, expected, actual)
}

func sortSchema(schema *messageflow.Schema) {
	sort.Slice(schema.Services, func(i, j int) bool {
		return schema.Services[i].Name < schema.Services[j].Name
	})

	for i := range schema.Services {
		sort.Slice(schema.Services[i].Operation, func(j, k int) bool {
			if schema.Services[i].Operation[j].Action != schema.Services[i].Operation[k].Action {
				return schema.Services[i].Operation[j].Action < schema.Services[i].Operation[k].Action
			}
			return schema.Services[i].Operation[j].Channel.Name < schema.Services[i].Operation[k].Channel.Name
		})
	}
}
