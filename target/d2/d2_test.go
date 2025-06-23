package d2

import (
	"context"
	_ "embed"
	"os"
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
							Message: messageflow.Message{
								Name: "PreferencesRequest",
								Payload: `{
  "user_id": "string[uuid]"
}`,
							},
						},
						Reply: &messageflow.Channel{
							Name: "notification.preferences.get",
							Message: messageflow.Message{
								Name: "PreferencesReply",
								Payload: `{
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
					},
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "notification.preferences.update",
							Message: messageflow.Message{
								Name: "PreferencesUpdate",
								Payload: `{
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
					},
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "notification.user.{user_id}.push",
							Message: messageflow.Message{
								Name: "PushNotification",
								Payload: `{
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
					},
					{
						Action: messageflow.ActionSend,
						Channel: messageflow.Channel{
							Name: "user.info.request",
							Message: messageflow.Message{
								Name:    "UserInfoRequest",
								Payload: `{"user_id": "string[uuid]"}`,
							},
						},
						Reply: &messageflow.Channel{
							Name: "user.info.request",
							Message: messageflow.Message{
								Name:    "UserInfoReply",
								Payload: `{"email": "string[email]", "name": "string"}`,
							},
						},
					},
					{
						Action: messageflow.ActionSend,
						Channel: messageflow.Channel{
							Name: "notification.analytics",
							Message: messageflow.Message{
								Name:    "AnalyticsEvent",
								Payload: `{"event_type": "string", "user_id": "string[uuid]"}`,
							},
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	target, err := NewTarget()
	require.NoError(t, err)

	actual, err := target.FormatSchema(ctx, schema, messageflow.FormatOptions{
		Mode: messageflow.FormatModeServiceChannels,
	})
	require.NoError(t, err)

	if os.Getenv("REWRITE_TESTDATA") == "true" {
		err = os.WriteFile("testdata/schema.d2", actual.Data, 0644)
		require.NoError(t, err)

		return
	}

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

	if os.Getenv("REWRITE_TESTDATA") == "true" {
		err = os.WriteFile("testdata/schema.svg", actual, 0644)
		require.NoError(t, err)

		return
	}

	assert.Equal(t, testSVG, actual)
}

func TestFormatSchemaChannelServicesWithOmitPayloads(t *testing.T) {
	t.Parallel()

	schema := messageflow.Schema{
		Services: []messageflow.Service{
			{
				Name: "Notification Service",
				Operation: []messageflow.Operation{
					{
						Action: messageflow.ActionSend,
						Channel: messageflow.Channel{
							Name: "notification.send",
							Message: messageflow.Message{
								Name:    "NotificationMessage",
								Payload: `{"message": "test"}`,
							},
						},
					},
				},
			},
			{
				Name: "User Service",
				Operation: []messageflow.Operation{
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "notification.send",
							Message: messageflow.Message{
								Name:    "NotificationMessage",
								Payload: `{"message": "test"}`,
							},
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	target, err := NewTarget()
	require.NoError(t, err)

	// Test with OmitPayloads = false (default behavior)
	actual, err := target.FormatSchema(ctx, schema, messageflow.FormatOptions{
		Mode:         messageflow.FormatModeChannelServices,
		Channel:      "notification.send",
		OmitPayloads: false,
	})
	require.NoError(t, err)

	// Should contain payload information
	actualStr := string(actual.Data)
	assert.Contains(t, actualStr, "Message(NotificationMessage):")
	assert.Contains(t, actualStr, `{"message": "test"}`)

	// Test with OmitPayloads = true
	actualOmitted, err := target.FormatSchema(ctx, schema, messageflow.FormatOptions{
		Mode:         messageflow.FormatModeChannelServices,
		Channel:      "notification.send",
		OmitPayloads: true,
	})
	require.NoError(t, err)

	// Should not contain payload information
	actualOmittedStr := string(actualOmitted.Data)
	assert.NotContains(t, actualOmittedStr, "Message(NotificationMessage):")
	assert.NotContains(t, actualOmittedStr, `{"message": "test"}`)
	assert.Contains(t, actualOmittedStr, "notification.send")
	assert.Contains(t, actualOmittedStr, "Notification Service")
	assert.Contains(t, actualOmittedStr, "User Service")
}

func TestFormatSchemaServiceServices(t *testing.T) {
	t.Parallel()

	schema := messageflow.Schema{
		Services: []messageflow.Service{
			{
				Name:        "Notification Service",
				Description: "Handles notification operations",
				Operation: []messageflow.Operation{
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "notification.preferences.get",
							Message: messageflow.Message{
								Name:    "PreferencesRequest",
								Payload: `{"user_id": "string[uuid]"}`,
							},
						},
						Reply: &messageflow.Channel{
							Name: "notification.preferences.get",
							Message: messageflow.Message{
								Name:    "PreferencesReply",
								Payload: `{"preferences": {"email_enabled": "boolean"}}`,
							},
						},
					},
					{
						Action: messageflow.ActionSend,
						Channel: messageflow.Channel{
							Name: "user.info.request",
							Message: messageflow.Message{
								Name:    "UserInfoRequest",
								Payload: `{"user_id": "string[uuid]"}`,
							},
						},
						Reply: &messageflow.Channel{
							Name: "user.info.request",
							Message: messageflow.Message{
								Name:    "UserInfoReply",
								Payload: `{"email": "string[email]", "name": "string"}`,
							},
						},
					},
					{
						Action: messageflow.ActionSend,
						Channel: messageflow.Channel{
							Name: "notification.analytics",
							Message: messageflow.Message{
								Name:    "AnalyticsEvent",
								Payload: `{"event_type": "string", "user_id": "string[uuid]"}`,
							},
						},
					},
				},
			},
			{
				Name:        "User Service",
				Description: "Manages user information",
				Operation: []messageflow.Operation{
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "user.info.request",
							Message: messageflow.Message{
								Name:    "UserInfoRequest",
								Payload: `{"user_id": "string[uuid]"}`,
							},
						},
						Reply: &messageflow.Channel{
							Name: "user.info.request",
							Message: messageflow.Message{
								Name:    "UserInfoReply",
								Payload: `{"email": "string[email]", "name": "string"}`,
							},
						},
					},
				},
			},
			{
				Name:        "Analytics Service",
				Description: "Tracks analytics events",
				Operation: []messageflow.Operation{
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "notification.analytics",
							Message: messageflow.Message{
								Name:    "AnalyticsEvent",
								Payload: `{"event_type": "string", "user_id": "string[uuid]"}`,
							},
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	target, err := NewTarget()
	require.NoError(t, err)

	actual, err := target.FormatSchema(ctx, schema, messageflow.FormatOptions{
		Mode:    messageflow.FormatModeServiceServices,
		Service: "Notification Service",
	})
	require.NoError(t, err)

	actualStr := string(actual.Data)
	assert.Contains(t, actualStr, "Notification Service")
	assert.Contains(t, actualStr, "User Service")
	assert.Contains(t, actualStr, "notification.preferences.get")
	assert.Contains(t, actualStr, "user.info.request")
	assert.Contains(t, actualStr, "notification.analytics")
	assert.Contains(t, actualStr, "Analytics Service")
}
