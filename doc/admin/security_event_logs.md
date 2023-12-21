# Security Event Logs
This guide goes into the details of Securit Event Logging in Sourcegraph
> Note: You can find more information about our audit logs setup [here](./audit_log.md)
>
> [Here](../dev/how-to/add_logging.md) is a guide on how to add logging to Sourcegraph backend

## What are Security Event Logs
- The purpose of Security Event Logs is to allow security specialists to be able to trace the steps of a user or an admin across the application.
- Getting a full picture, of how a user moves through the application, in a single location is crucial for many reasons.
- When a user takes an action on a sensitive part of the application, this should be logged to make sure it can be retraced to a user and time.
- In Sourcegraph application, we log these sensitive actions as "security event" with relevant information included in the output.
- These logs can be enabled/disabled as well the location can be set via the [site config settings](./audit_log#configuring)


## How to log a security event
- All the logging for security event is done through our security_event_log.go functions
- To log an event, the LogSecurityEvent function can be invoked which will create an event with information provided and then submit it for logging to the right place
- This function takes following information to create a log event
  - Context from the where the log is being called
  - SecurityEventName which is predefined here
  - URL if available
  - userID of the user that the action is applied towards
  - anonymousUserID for unauthenitcated users
  - source of the log
  - arguments relevant to the action
- Example of invoking the function
  ```go
  db.SecurityEventLogs().LogSecurityEvent(ctx, database.SecurityEventNameEmailVerified, r.URL.Path, uint32(actr.UID), "", "BACKEND", email)

- The function sends the log event it creates to be pushed to the right location based on the site-config settings
- The function also checks to make sure that marshalling the arguments does not cause as error

## How to find security events in logs


