Code Review Checklist Details
=============================

This document provides a detailed descriptions of items in [`PULL_REQUEST_TEMPLATE.md`](PULL_REQUEST_TEMPLATE.md).

<a id="ops">
Ops
---

#### Monitoring & alerts

- Pingdom (connected to monitoring-bot and OpsGenie)
- Prometheus (connected to OpsGenie)
- Splunk alerts (connected to Slack & OpsGenie)
- Monitoring-bot (auto-rollbacks)
- End-to-end Selenium tests
- k8s health check on start-up
- PostgreSQL alerts
- Sentry
- Sysdig

#### Tracing

- LightStep

#### User tracking

- Amplitude Analytics
- Mode Analytics
- Google Analytics
- Intercom
- FullStory

Explicit omissions
------------------

- Public API: we currently do not support backcompat against our API. This will change when we publish it.
- Data migration: this should be obvious enough to the person making the change, and given that we have continuous deployment, migration issues should be detected and rolled back pretty quickly.
