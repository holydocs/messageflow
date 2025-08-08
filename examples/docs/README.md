# Message Flow

## Table of Contents

- [Context](#context)
- [Services](#services)
  - [Analytics Service](#analytics-service)
  - [Campaign Service](#campaign-service)
  - [Notification Service](#notification-service)
  - [User Service](#user-service)
- [Channels](#channels)
  - [analytics.insights](#analyticsinsights)
  - [analytics.report.request](#analyticsreportrequest)
  - [analytics.warning](#analyticswarning)
  - [campaign.analytics](#campaignanalytics)
  - [campaign.create](#campaigncreate)
  - [campaign.execute](#campaignexecute)
  - [campaign.status](#campaignstatus)
  - [notification.analytics](#notificationanalytics)
  - [notification.preferences.get](#notificationpreferencesget)
  - [notification.preferences.update](#notificationpreferencesupdate)
  - [notification.user.{user_id}.push](#notificationuseruser_idpush)
  - [user.analytics](#useranalytics)
  - [user.info.request](#userinforequest)
  - [user.info.update](#userinfoupdate)
- [Changelog](#changelog)

## Context

![Context](diagrams/context.svg)

## Services

### Analytics Service

A centralized analytics service that receives and processes analytics events from all other services.
Provides insights, reporting, and analytics data aggregation for user behavior, notification performance,
campaign effectiveness, and system-wide metrics.


![Analytics Service Service Channels](diagrams/service_analytics-service.svg)

### Campaign Service

A service that manages notification campaigns, user targeting, and campaign execution.
Handles campaign creation, user segmentation, scheduling, and personalized notification delivery.
Uses user data for targeting and personalization of campaign messages.


![Campaign Service Service Channels](diagrams/service_campaign-service.svg)

### Notification Service

A service that handles user notifications, preferences, and interactions.
Supports real-time notifications, user preferences management.


![Notification Service Service Channels](diagrams/service_notification-service.svg)

### User Service

A service that manages user information, profiles, and authentication.
Handles user data requests, profile updates, and user lifecycle events.


![User Service Service Channels](diagrams/service_user-service.svg)

## Channels

### analytics.insights

![analytics.insights Channel Services](diagrams/channel_analyticsinsights.svg)

#### Messages
**AnalyticsInsightMessage**
```json
{
  "category": "string[enum:user_behavior,notification_performance,campaign_effectiveness,system_health]",
  "confidence": "number[float]",
  "created_at": "string[date-time]",
  "data_points": [
    "object"
  ],
  "description": "string",
  "insight_id": "string[uuid]",
  "insight_type": "string[enum:trend,anomaly,recommendation,warning]",
  "metadata": {
    "environment": "string[enum:development,staging,production]",
    "platform": "string[enum:ios,android,web]",
    "source": "string[enum:mobile,web,api]",
    "version": "string"
  },
  "recommendations": [
    "string"
  ],
  "severity": "string[enum:low,medium,high,critical]",
  "title": "string"
}
```

### analytics.report.request

![analytics.report.request Channel Services](diagrams/channel_analyticsreportrequest.svg)

#### Messages
**request**: AnalyticsReportRequestMessage
```json
{
  "created_at": "string[date-time]",
  "filters": {
    "campaign_ids": [
      "string[uuid]"
    ],
    "event_types": [
      "string"
    ],
    "user_ids": [
      "string[uuid]"
    ],
    "user_segments": [
      "string[enum:all_users,new_users,active_users,inactive_users,premium_users,free_users]"
    ]
  },
  "format": "string[enum:json,csv,pdf]",
  "metrics": [
    "string[enum:event_count,user_count,conversion_rate,engagement_rate,response_time,error_rate]"
  ],
  "report_id": "string[uuid]",
  "report_type": "string[enum:user_activity,notification_performance,campaign_effectiveness,system_health,custom]",
  "time_range": {
    "end": "string[date-time]",
    "granularity": "string[enum:minute,hour,day,week,month]",
    "start": "string[date-time]"
  }
}
```
**reply**: AnalyticsReportReplyMessage
```json
{
  "data": "object",
  "error": {
    "code": "string",
    "message": "string"
  },
  "generated_at": "string[date-time]",
  "insights": [
    {
      "confidence": "number[float]",
      "data_points": [
        "object"
      ],
      "description": "string",
      "impact": "string[enum:low,medium,high]",
      "title": "string",
      "type": "string[enum:trend,anomaly,correlation,recommendation]"
    }
  ],
  "report_id": "string[uuid]",
  "report_type": "string[enum:user_activity,notification_performance,campaign_effectiveness,custom]",
  "summary": {
    "event_types": "object",
    "top_metrics": {
      "conversion_rate": "number[float]",
      "engagement_rate": "number[float]",
      "error_rate": "number[float]",
      "response_time_avg": "number[float]"
    },
    "total_events": "integer",
    "unique_users": "integer"
  },
  "time_range": {
    "end": "string[date-time]",
    "granularity": "string[enum:minute,hour,day,week,month]",
    "start": "string[date-time]"
  }
}
```

### analytics.warning

![analytics.warning Channel Services](diagrams/channel_analyticswarning.svg)

#### Messages
**AnalyticsWarningMessage**
```json
{
  "actions": [
    "string"
  ],
  "affected_services": [
    "string[enum:user_service,notification_service,campaign_service]"
  ],
  "created_at": "string[date-time]",
  "current_value": "number",
  "description": "string",
  "metadata": {
    "environment": "string[enum:development,staging,production]",
    "platform": "string[enum:ios,android,web]",
    "source": "string[enum:mobile,web,api]",
    "version": "string"
  },
  "metric": "string",
  "severity": "string[enum:low,medium,high,critical]",
  "threshold": "number",
  "time_window": "string",
  "title": "string",
  "warning_id": "string[uuid]",
  "warning_type": "string[enum:anomaly_detected,threshold_exceeded,trend_change,system_issue]"
}
```

### campaign.analytics

![campaign.analytics Channel Services](diagrams/channel_campaignanalytics.svg)

#### Messages
**CampaignAnalyticsEventMessage**
```json
{
  "campaign_id": "string[uuid]",
  "event_id": "string[uuid]",
  "event_type": "string[enum:campaign_created,campaign_executed,notification_sent,notification_opened,notification_clicked,campaign_completed,campaign_failed]",
  "execution_id": "string[uuid]",
  "metadata": {
    "environment": "string[enum:development,staging,production]",
    "platform": "string[enum:ios,android,web]",
    "source": "string[enum:mobile,web,api]",
    "version": "string"
  },
  "notification_id": "string[uuid]",
  "timestamp": "string[date-time]",
  "user_id": "string[uuid]"
}
```

### campaign.create

![campaign.create Channel Services](diagrams/channel_campaigncreate.svg)

#### Messages
**CampaignCreateMessage**
```json
{
  "campaign_id": "string[uuid]",
  "created_at": "string[date-time]",
  "description": "string",
  "metadata": {
    "environment": "string[enum:development,staging,production]",
    "platform": "string[enum:ios,android,web]",
    "source": "string[enum:mobile,web,api]",
    "version": "string"
  },
  "name": "string",
  "notification_template": {
    "body_template": "string",
    "data": "object",
    "localization": "object",
    "priority": "string[enum:low,normal,high]",
    "title_template": "string"
  },
  "schedule": {
    "recurring": {
      "end_date": "string[date]",
      "frequency": "string[enum:daily,weekly,monthly]",
      "interval": "integer",
      "start_date": "string[date]"
    },
    "scheduled_at": "string[date-time]",
    "timezone": "string",
    "type": "string[enum:immediate,scheduled,recurring]"
  },
  "settings": {
    "a_b_testing": {
      "enabled": "boolean",
      "traffic_split": [
        "number"
      ],
      "variants": [
        {
          "body_template": "string",
          "data": "object",
          "localization": "object",
          "priority": "string[enum:low,normal,high]",
          "title_template": "string"
        }
      ]
    },
    "batch_size": "integer",
    "max_retries": "integer",
    "rate_limit": "integer",
    "respect_quiet_hours": "boolean"
  },
  "target_audience": {
    "estimated_reach": "integer",
    "user_filters": {
      "language": [
        "string"
      ],
      "last_activity": {
        "from": "string[date-time]",
        "to": "string[date-time]"
      },
      "registration_date": {
        "from": "string[date]",
        "to": "string[date]"
      },
      "timezone": [
        "string"
      ]
    },
    "user_segments": [
      "string[enum:all_users,new_users,active_users,inactive_users,premium_users,free_users]"
    ]
  }
}
```

### campaign.execute

![campaign.execute Channel Services](diagrams/channel_campaignexecute.svg)

#### Messages
**CampaignExecuteMessage**
```json
{
  "batch_size": "integer",
  "campaign_id": "string[uuid]",
  "created_at": "string[date-time]",
  "execution_id": "string[uuid]",
  "execution_type": "string[enum:immediate,scheduled,batch]",
  "metadata": {
    "environment": "string[enum:development,staging,production]",
    "platform": "string[enum:ios,android,web]",
    "source": "string[enum:mobile,web,api]",
    "version": "string"
  },
  "priority": "string[enum:low,normal,high]"
}
```

### campaign.status

![campaign.status Channel Services](diagrams/channel_campaignstatus.svg)

#### Messages
**CampaignStatusUpdateMessage**
```json
{
  "campaign_id": "string[uuid]",
  "error": {
    "code": "string",
    "message": "string"
  },
  "execution_id": "string[uuid]",
  "progress": {
    "failed": "integer",
    "sent": "integer",
    "success_rate": "number[float]",
    "total_targets": "integer"
  },
  "status": "string[enum:pending,running,completed,failed,paused,cancelled]",
  "updated_at": "string[date-time]"
}
```

### notification.analytics

![notification.analytics Channel Services](diagrams/channel_notificationanalytics.svg)

#### Messages
**NotificationAnalyticsEventMessage**
```json
{
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
}
```

### notification.preferences.get

![notification.preferences.get Channel Services](diagrams/channel_notificationpreferencesget.svg)

#### Messages
**request**: PreferencesRequestMessage
```json
{
  "user_id": "string[uuid]"
}
```
**reply**: PreferencesReplyMessage
```json
{
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
}
```

### notification.preferences.update

![notification.preferences.update Channel Services](diagrams/channel_notificationpreferencesupdate.svg)

#### Messages
**PreferencesUpdateMessage**
```json
{
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
}
```

### notification.user.{user_id}.push

![notification.user.{user_id}.push Channel Services](diagrams/channel_notificationuseruser_idpush.svg)

#### Messages
**PushNotificationMessage**
```json
{
  "body": "string",
  "created_at": "string[date-time]",
  "data": "object",
  "notification_id": "string[uuid]",
  "priority": "string[enum:low,normal,high]",
  "title": "string",
  "user_id": "string[uuid]"
}
```

### user.analytics

![user.analytics Channel Services](diagrams/channel_useranalytics.svg)

#### Messages
**UserAnalyticsEventMessage**
```json
{
  "event_id": "string[uuid]",
  "event_type": "string[enum:user_registered,user_logged_in,profile_updated,preferences_changed,account_deleted]",
  "metadata": {
    "environment": "string[enum:development,staging,production]",
    "platform": "string[enum:ios,android,web]",
    "source": "string[enum:mobile,web,api]",
    "version": "string"
  },
  "timestamp": "string[date-time]",
  "user_id": "string[uuid]"
}
```

### user.info.request

![user.info.request Channel Services](diagrams/channel_userinforequest.svg)

#### Messages
**request**: UserInfoRequestMessage
```json
{
  "user_id": "string[uuid]"
}
```
**reply**: UserInfoReplyMessage
```json
{
  "email": "string[email]",
  "error": {
    "code": "string",
    "message": "string"
  },
  "language": "string",
  "name": "string",
  "timezone": "string",
  "user_id": "string[uuid]"
}
```

### user.info.update

![user.info.update Channel Services](diagrams/channel_userinfoupdate.svg)

#### Messages
**UserInfoUpdateMessage**
```json
{
  "changes": "object",
  "metadata": {
    "environment": "string[enum:development,staging,production]",
    "platform": "string[enum:ios,android,web]",
    "source": "string[enum:mobile,web,api]",
    "version": "string"
  },
  "updated_at": "string[date-time]",
  "user_id": "string[uuid]"
}
```

## Changelog

### 2025-08-08
- **added** channel: 'send' on channel 'analytics.warning' was added to service 'Analytics Service'
- **changed** message: Reply message payload changed for operation 'receive' on channel 'analytics.report.request' in service 'Analytics Service'
```json
  (
  	"""
  	... // 18 identical lines
  	  ],
  	  "report_id": "string[uuid]",
- 	  "report_type": "string[enum:user_activity,notification_performance,campaign_effectiveness,system_health,custom]",
+ 	  "report_type": "string[enum:user_activity,notification_performance,campaign_effectiveness,custom]",
  	  "summary": {
  	    "event_types": "object",
  	... // 15 identical lines
  	"""
  )

```
- **changed** message: Message payload changed for operation 'send' on channel 'analytics.insights' in service 'Analytics Service'
```json
  (
  	"""
  	... // 7 identical lines
  	  "description": "string",
  	  "insight_id": "string[uuid]",
- 	  "insight_type": "string[enum:trend,anomaly,recommendation,alert]",
+ 	  "insight_type": "string[enum:trend,anomaly,recommendation,warning]",
  	  "metadata": {
  	    "environment": "string[enum:development,staging,production]",
  	... // 10 identical lines
  	"""
  )

```
- **removed** channel: 'send' on channel 'analytics.alert' was removed from service 'Analytics Service'

### 2025-08-08
- **added** service: 'User Service' was added
