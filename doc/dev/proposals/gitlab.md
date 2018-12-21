Direct GitLab integration v1
============================

Build Sourcegraph extensions directly into GitLab to provide code
intelligence and extensibility.

Background
----------

We want Sourcegraph to be built into GitLab so that everyone who uses
the code host will have it automatically without needing to install the
browser extension. [There is an issue open on GitLab for
this.](https://gitlab.com/gitlab-org/gitlab-ce/issues/41925)

Proposal
--------

We will build Sourcegraph extension support into GitLab. The integration
will point to Sourcegraph.com's extension registry.

GitLab will provide the
[CodeHost](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/browser/src/libs/code_intelligence/code_intelligence.tsx#L112:8)
implementation and Sourcegraph will serve an abstracted bundle similar
to Phabricator that takes GitLab's code host implementation and uses it
to register extension support. We are separating ownership of the code
host and the script so that GitLab owns the CodeHost implementation and
we own the handling of it. This logical split also brings along the
benefit of other code hosts being able to use this script without any
changes required to Sourcegraph.

### Test plan

We will add tests to GitLab to ensure that any breaking changes to the
DOM will require updates to the CodeHost. Our current and future tests
in the browser extension will test the handling of the code host and
code intelligence.

### Release plan

Releasing this will be a two part process:

1.  Once merged into GitLab, we're on their release schedule.
2.  Releasing the script we use to provide code intelligence will be
    similar to Phabricator. We'll just serve it as an asset from
    Sourcegraph.

Rationale
---------

We want increased distrubution and GitLab wants code intelligence and
extensibility of their product. There are unknowns:

1.  How long will the code review process take?
2.  Will GitLab accept the integration while it points at
    Sourcegraph.com for extensions?

Implementation
--------------

- [x] Spec out how we are going to include JS in GitLab
- [x] Open up WIP MR on GitLab
  ([!23755](https://gitlab.com/gitlab-org/gitlab-ce/merge_requests/23755))
- [ ] Impliment UI for configuring Sourcegraph extensions
  - [ ] Mocks
- [ ] Implement CodeHost interface in GitLab
- [ ] Remove CORS checks around extensions so that GitLab can fetch
  extensions. This is all public data anyways.
- [1/2] Script that runs extensions using acces token and CodeHost
  from GitLab
  - [x] Create script
  - [ ] Modify it so that it works with GitLab
- Future works
  - Implement Sourcegraph extension to provide code intelligence for
    GitLab specific files such as `gitlab.yml` for CI.
