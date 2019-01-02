# Direct GitLab integration v1

Build Sourcegraph extensions directly into GitLab to provide code
intelligence and extensibility.

## Background

We want Sourcegraph build Sourcegraph extension support directly into GitLab so that everyone who uses
the code host will have it automatically without needing to install the
browser extension. [There is an issue open on GitLab for
this.](https://gitlab.com/gitlab-org/gitlab-ce/issues/41925)

## Plan

We will build Sourcegraph extension support into GitLab. The integration
will point to Sourcegraph.com's extension registry. This means that GitLab users
will be able to use any extension that is available at [sourcegraph.com/extensions](https://sourcegraph.com/extensions).

The browser extension supports different code hosts. It does this by
implementing the Sourcegraph specific logic in an abstract way. It consumes
implementations of a [`CodeHost`
interface](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/browser/src/libs/code_intelligence/code_intelligence.tsx#L112:8)
which each supported code host (e.g. GitHub, GitLab, etc) is required to
implement. This interface tells the abstracted logic how to look at the DOM
within the context that its running in to get the information needed (e.g. Repo name,
revision, file name, etc). This allows us to implement support for
new features once and, as long as the code host provides what we need to know,
the new feature works for all code hosts.

In this project, GitLab will provide the `CodeHost`
implementation and Sourcegraph will serve a script that picks up the
`CodeHost` implementation from the `window` object and uses it
to register extension support. This script is the same as the browser
extension's code sans code host detection logic. We are separating ownership of the code
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

### Success metrics

We gain distribution by each GitLab user being exposed to Sourcegraph extensions
out of the box.

### Company goals

This aligns with our company goals because it increases distribution by putting
Sourcegraph in front of every GitLab user.

## Rationale

We want increased distribution and GitLab wants code intelligence and
extensibility of their product. There are unknowns:

1.  How long will the code review process take?
2.  Will GitLab accept the integration while it points at
    Sourcegraph.com for extensions?

## Checklist

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

## Done date

03/01/2019 (3.1 release)

## Retrospective

[This section is completed after the project is completely done (i.e. the checklist is complete).]

### Actual checklist

[The checklist that was actually completed (i.e. paste the final checklist from the issue). Explain any differences from the original checklist in the plan.]

### Actual done date

[The date that the project was actually finished. Explain why this is earlier or
later than originally planned or explain why the project was not completed.]
