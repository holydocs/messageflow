package d2

import (
	"context"
	"os"
	"testing"

	"github.com/denchenko/messageflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			{
				Name: "User Service",
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
					{
						Action: messageflow.ActionSend,
						Channel: messageflow.Channel{
							Name: "user.created",
							Message: messageflow.Message{
								Name:    "UserCreated",
								Payload: `{"user_id": "string[uuid]", "email": "string[email]"}`,
							},
						},
					},
				},
			},
			{
				Name: "Analytics Service",
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
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "user.created",
							Message: messageflow.Message{
								Name:    "UserCreated",
								Payload: `{"user_id": "string[uuid]", "email": "string[email]"}`,
							},
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		opts         messageflow.FormatOptions
		description  string
		testdataFile string
	}{
		{
			name: "FormatModeServiceChannels - Notification Service",
			opts: messageflow.FormatOptions{
				Mode:    messageflow.FormatModeServiceChannels,
				Service: "Notification Service",
			},
			description:  "Should format schema showing channels for Notification Service",
			testdataFile: "service_channels_notification.d2",
		},
		{
			name: "FormatModeChannelServices - notification.analytics",
			opts: messageflow.FormatOptions{
				Mode:         messageflow.FormatModeChannelServices,
				Channel:      "notification.analytics",
				OmitPayloads: false,
			},
			description:  "Should format schema showing services for notification.analytics channel",
			testdataFile: "channel_services_notification_analytics.d2",
		},
		{
			name: "FormatModeContextServices",
			opts: messageflow.FormatOptions{
				Mode: messageflow.FormatModeContextServices,
			},
			description:  "Should format schema showing context view of all services and their connections",
			testdataFile: "context_services.d2",
		},
		{
			name: "FormatModeServiceServices - User Service",
			opts: messageflow.FormatOptions{
				Mode:    messageflow.FormatModeServiceServices,
				Service: "User Service",
			},
			description:  "Should format schema showing User Service and its neighboring services",
			testdataFile: "service_services_user.d2",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target, err := NewTarget()
			require.NoError(t, err)

			actual, err := target.FormatSchema(ctx, schema, tt.opts)
			require.NoError(t, err)
			require.NotNil(t, actual)

			assert.Equal(t, messageflow.TargetType("d2"), actual.Type)

			if os.Getenv("OVERWRITE_TESTDATA") == "true" {
				err = os.WriteFile("testdata/"+tt.testdataFile, actual.Data, 0644)
				require.NoError(t, err)
				return
			}

			expectedData, err := os.ReadFile("testdata/" + tt.testdataFile)
			require.NoError(t, err)

			expected := messageflow.FormattedSchema{
				Type: messageflow.TargetType("d2"),
				Data: expectedData,
			}
			assert.Equal(t, expected, actual)
		})
	}
}

func TestRenderSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

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
			{
				Name: "User Service",
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
					{
						Action: messageflow.ActionSend,
						Channel: messageflow.Channel{
							Name: "user.created",
							Message: messageflow.Message{
								Name:    "UserCreated",
								Payload: `{"user_id": "string[uuid]", "email": "string[email]"}`,
							},
						},
					},
				},
			},
			{
				Name: "Analytics Service",
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
					{
						Action: messageflow.ActionReceive,
						Channel: messageflow.Channel{
							Name: "user.created",
							Message: messageflow.Message{
								Name:    "UserCreated",
								Payload: `{"user_id": "string[uuid]", "email": "string[email]"}`,
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name         string
		opts         messageflow.FormatOptions
		description  string
		testdataFile string
	}{
		{
			name: "FormatModeServiceChannels - Notification Service",
			opts: messageflow.FormatOptions{
				Mode:    messageflow.FormatModeServiceChannels,
				Service: "Notification Service",
			},
			description:  "Should render schema showing channels for Notification Service",
			testdataFile: "service_channels_notification.svg",
		},
		{
			name: "FormatModeChannelServices - notification.analytics",
			opts: messageflow.FormatOptions{
				Mode:         messageflow.FormatModeChannelServices,
				Channel:      "notification.analytics",
				OmitPayloads: false,
			},
			description:  "Should render schema showing services for notification.analytics channel",
			testdataFile: "channel_services_notification_analytics.svg",
		},
		{
			name: "FormatModeContextServices",
			opts: messageflow.FormatOptions{
				Mode: messageflow.FormatModeContextServices,
			},
			description:  "Should render schema showing context view of all services and their connections",
			testdataFile: "context_services.svg",
		},
		{
			name: "FormatModeServiceServices - User Service",
			opts: messageflow.FormatOptions{
				Mode:    messageflow.FormatModeServiceServices,
				Service: "User Service",
			},
			description:  "Should render schema showing User Service and its neighboring services",
			testdataFile: "service_services_user.svg",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target, err := NewTarget()
			require.NoError(t, err)

			formattedSchema, err := target.FormatSchema(ctx, schema, tt.opts)
			require.NoError(t, err)
			require.NotNil(t, formattedSchema)

			actual, err := target.RenderSchema(ctx, formattedSchema)
			require.NoError(t, err)
			require.NotNil(t, actual)

			if os.Getenv("OVERWRITE_TESTDATA") == "true" {
				err = os.WriteFile("testdata/"+tt.testdataFile, actual, 0644)
				require.NoError(t, err)
				return
			}

			expectedData, err := os.ReadFile("testdata/" + tt.testdataFile)
			require.NoError(t, err)

			assert.Equal(t, expectedData, actual)
		})
	}
}
