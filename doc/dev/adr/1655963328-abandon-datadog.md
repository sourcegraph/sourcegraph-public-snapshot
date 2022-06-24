# 4. Sunset Datadog

Date: 2022-06-23

### Context

RFC 575 led to implement a Datadog POC, in an effort to increase and unify the observability tooling. 

Reasons to not got further with the POC: 

- We have been underutilizing the platform quite heavily.
- The DevOps team no longer exists so we would need to find a new owner for this service soon.
- Unfavourable consensus around its usability.
- Moving toward MI as our main priority.

### Decision

Revert to use Loki as logging aggregator and HoneyComb for tracing purposes.

### Consequences

- https://github.com/sourcegraph/sourcegraph/issues/37568
- DevX team is responsible of sunsetting it with Dax's guidance.
