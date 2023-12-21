# Security Event Logs
This guide goes into the details of Securit Event Logging in Sourcegraph
> Note: You can find more information about our audit logs setup [here](./audit_log.md)
>
> [Here](../dev/how-to/add_logging.md) is a guide on how to add logging to Sourcegraph backend

## What are Security Event Logs
- The purpose of Security Event Logs is to allow security specialists to be able to trace the steps of a user or an admin across the application.
- Getting a full picture, of how a user moves through the application, in a single location is crucial for many reasons.
- When a user takes an action on a sensitive information within the application, this should be logged to make sure it can be retraced to the user and time.
- In Sourcegraph application, these sensitive actions are logged as "security event" with relevant information included in the output.
- These logs can be enabled/disabled as well as the location can be set via the [site config settings](./audit_log#configuring)
- Previously, we were logging very selective set of actions. However, through various analyses, it was determined that there were some gaps in creating the full picture.
  - New event types are constantly being added to fill these gaps.


## How to log a security event
- All the logging for security event is done through our [security_event_log.go](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/security_event_logs.go) functions
- Previously, events were created within the function where the action was taking place and then pushed to the logging location
  ```go
  event := &SecurityEvent{
		Name:            SecurityEventNameAccessGranted,
		URL:             "",
		UserID:          uint32(a.UID),
		AnonymousUserID: "",
		Argument:        arg,
		Source:          "BACKEND",
		Timestamp:       time.Now(),
	}

	db.SecurityEventLogs().LogEvent(ctx, event)
  
- With a recent change to streamline the process, to log an event, the [LogSecurityEvent](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/security_event_logs.go?L253:34&popover=pinned) function can be invoked which takes care of marshalling the arguments and creating the SecurityEvent.
- This function takes following information to create a log event
  - Context contains information on the acting user
  - SecurityEventName which is predefined [here](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/security_event_logs.go?L22-101)
  - URL if available
  - userID of the user that the action is applied towards
  - anonymousUserID for unauthenitcated users
  - source of the log
  - arguments relevant to the action being logged
- Example of using the function to logan event
  ```go
  db.SecurityEventLogs().LogSecurityEvent(ctx, database.SecurityEventNameEmailAdded, r.URL.Path, uint32(actr.UID), "", "BACKEND", email)

- The function sends the log event it creates to be pushed to the right location based on the site-config settings
- The function also checks to make sure that marshalling the arguments does not cause as error

## How to find security events in logs
- Security events are logged with all the relevant information associated with the actions
- Depending on the location of the log destination, the event log can be either found in the application log output or in the database or both.
- A sample output of a logged event from application logs would look similar to this:
  ```JSON
  {
  ...
  {
    "message": "EmailAdded (sampling immunity token: 12345-222-3333-5454-9w08fs0s9d8f)",
    "Caller": "audit/audit.go:57",
    "Attributes": {
      "audit": {
        "actor": {
          "X-Forwarded-For": "127.0.0.1",
          "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) ",
          "ip": "100.211.3.155",
          "actorUID": "123245"
        },
        "entity": "security events",
        "auditId": "12345-222-3333-5454-9w08fs0s9d8f",
        "action": "EmailAdded"
      },
      "event": {
        "URL": "",
        "argument": "\"new@sourcegraph.com\"",
        "AnonymousUserID": "",
        "UserID": 223955,
        "source": "12345",
        "timestamp": "2023-12-21 02:41:08.649603776 +0000 UTC",
        "version": "255367_2023-12-20_5.2-a3143120c41e"
      }
    },
    "Function": "github.com/sourcegraph/sourcegraph/internal/audit.Log",
    "InstrumentationScope": "frontend.SecurityEvents",
    "timestampNanos": 1703126468649641000,
    "Resource": {
      "service.instance.id": "sourcegraph-frontend-769bdbdd77-p2f8j",
      "service.name": "frontend",
      "service.version": "255367_2023-12-20_5.2-a3143120c41e"
    }
  }
  ...
  }

- Entity field can be used to filter on all security events.
- Action field will provide information on the event and can be correlated with the action taken, in this case EmailAdded
- The actorUID can be used to filter out events from a particular user
- UserID can be used to filter out actions taken on a particular user's information
  
## FAQ
What events are currently being logged as security events?
- [These](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/security_event_logs.go?L22-101) are the events that are currently being logged.

What if I dont want these events to be logged?
- To turn off all security event logs, you can [set the variable](https://docs.sourcegraph.com/admin/audit_log#excessive-audit-logging) in the site config

How can I correlate the actorID or userID to a user in the application?
- This correlation can be done by a site-admin via [graphql query in the API Console](https://docs.sourcegraph.com/admin/audit_log#faq).
