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
To log a security event, 

## How to find security events in logs


