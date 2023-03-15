# 8. Audit log on top of structured logging

Date: 2022-11-02

## Context

Sourcegraph has lacked first-class support for audit logging until now. This ADR doesn't intend to explain the business decisions behind building an audit log but the technical solution. For reading more about the business context, refer to the following RFCs:

- [RFC 732](https://docs.google.com/document/d/12x09FhwKz4hS-r__kPqn8k5pcnhJ_Iy3DHt7p6OKQcY/edit#heading=h.1yfwziyx4k3t)
- [RFC 734](https://docs.google.com/document/d/1tv1wNxfqtg4qfJ6jhbd9rFMIt1JCEjGEFzvV5j_i-sI/edit#heading=h.trqab8y0kufp)

For using and configuring the audit log, refer to the [Docs page](../../admin/audit_log.md)

## Decision

Add easy-to-use audit logging API, available in `internal/audit/audit.go`. The API entry point is the `audit.Log` function.

- Audit logs are regular structured logs that carry additional information in the `Attributes` map.
- Audit data is available in the `Attributes.audit` property.
- Any additional audit-related context is a direct child of the `Attributes` property.
- Audit logs are "immune" to sampling (they never get dropped); this is achieved using a unique message in the `Body` property (which carries a generated `auditId` UUID).

## Consequences

Audit log entries are now a part of the standard log output.

- On-premises installation may filter them from the stdout, looking for `Attributes.audit` or `auditId` properties.
- Cloud instances will ship with first-class support (TBD).
