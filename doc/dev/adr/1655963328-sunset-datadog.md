# 4. Sunset Datadog integration

Date: 2022-06-23

## Context

[RFC 575](https://docs.google.com/document/d/1xnAgloZB8sEkyhecjml2ByQl-aUCrJdWDYOBj3asA9g/edit) led to implement a Datadog POC, in an effort to decrease the proliferation of and unify the observability tooling used for operating Sourcegraph instances _within Sourcegraph_.

Reasons to not got further with the POC:

- We have been underutilizing the platform quite heavily
- The DevOps team no longer exists so we would need to find a new owner for this service soon.
- Unfavourable consensus around its usability.
  - _datadog's log search/viewer is really clunky and just doesn’t do the job well for me_ 
  - _you can't see fields alongside a log message, and the UI being a complete disaster_ 
- Moving toward MI as our main priority.

GCP Logs were envisioned but there was a strong push back, as its usuabilty was judged quite bad by teammates who worked with it.

## Decision

Sunset Datatog and drop all related code.

## Consequences

- https://github.com/sourcegraph/sourcegraph/issues/37568
- DevX team is responsible of sunsetting it with Dax's guidance.
