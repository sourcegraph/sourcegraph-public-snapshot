# Event level data usage pipeline

This document outlines information about the ability to export raw user event data from Sourcegraph. This is limited
to certain mangaged instances (cloud) where the customer has signed a corresponding data collection agreement.

### What is it?

This process is a background that can be enabled that will periodically scrape the `event_logs` table in the primary database
and send it to Sourcegraph centralized analytics. More information can be found in [RFC 719: Managed Instance Telemetry](https://docs.google.com/document/d/1N9aO0uTlvwXI7FzdPjIUCn_d1tRkJUfsc0urWigRf6s/edit).


### How to enable for a managed instance
1. Ensure the managed instance has the appropriate IAM policy applied (todo add link)
2. Update the managed instance deployment configuration to include the following environment variables:
3. `EXPORT_USAGE_DATA_ENABLED=true`
4. [`EXPORT_USAGE_DATA_TOPIC_NAME`](https://sourcegraph.sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/deploy-sourcegraph-managed%24+EXPORT_USAGE_DATA_TOPIC_NAME&patternType=standard)
5. [`EXPORT_USAGE_DATA_TOPIC_PROJECT`](https://sourcegraph.sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/deploy-sourcegraph-managed%24+EXPORT_USAGE_DATA_TOPIC_PROJECT&patternType=standard)
6. Deploy the updated deployment manifest and restart the `worker` service.

