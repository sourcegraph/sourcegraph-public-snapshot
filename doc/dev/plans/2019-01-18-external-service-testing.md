# External service testing

Create an easier and more reproducible way to run end to end tests against all
supported code hosts.

## Background

The browser extension is a huge part of our platform. It brings Sourcegraph
to several code hosts. More and more people continue to use it every day, an
they all expect it to work as expected every day. Currently, we only have a
few unit tests, and they're only run on GitHub.com. We should ensure greater
reliability and stability by having a more robust E2E test suite that covers more
cases.

In addition to greater reliability, a more robust E2E test suite would improve
developer experience. The test suite I'll outline below will provide developers
a way to spin up any code host into a known state easily so that they can easily
reproduce the bug and then write a test for the solution.

## Plan

To create a more robust end to end framework, we should make it easier to get
each code host into a known state. I plan on creating a tool that provides an
API for this. API includes the following "code host functions":

- Initialize
- Add repository
- Create code review (code review = GH/PR = GL/MR = Phab/Diff)
- Update code review
- Complete code review
- View code review
- View repo/rev/file
- Get token at position for code view of type (diff|single file)
- Shutdown

Once we know how to do this for the different code hosts, we know all the
required actions to get the code host into a state that we can run unit tests
against. We'll use this information to create test case repos and code reviews.
We'll have a test repo with a few files. We'll copy this directory multiple
times and incrementally change the files in each directory. These changes will
be used to simulate different commits and changes over time. The changes will
include various test cases that we'd like to test against such as multi line
changes, file name changes, etc.

To create the tree of commits, we'll replace the working directory with the next
directory and make a commit, then apply the change to the code host's code
review mechanism with the functions we've implemented.

The next part of the plan is to abstract the unit tests so they work with other
code hosts. Currently, they only work with GitHub.

To do this, we'll use the functions for viewing certain files or code reviews
for each code host. We'll modify the existing tests to point to the test case
repos we've created. In addition to this, the tests will have to be modified
to not use GH specific class names.

Finally, we'll ensure that they are able to run in CI. Currently, our E2E tests
finish and then hang in CI so they are disabled.

### Test plan

These are the tests :). We'll make sure they work by ensuring that we catch
failures by causing them on purpose and then fixing them.

### Release plan

Merge this in and hook them up to CI.

### Success metrics

This plan will be successful if we catch bugs before they are merged and
released or catch bugs that are caused by code hosts changing quickly.

### Company goals

This plan advances our company goals by helping us provide a more stable
product, which will keep our customers happy and encourage continued use.

## Rationale

[A discussion of alternate approaches and the trade offs, advantages,
disadvantages, risks, and uncertainties of the specified approach.]

Our current process for testing our external service integration with the
browser extension is lacking. We have E2E tests that are disabled in CI because
the process won't close. This means that developers are tasked with running them
on their own before merging which no one does. Even I only occasionally run
them. The reason for this is that they only run against GitHub and cover only a few
test cases.

To test other external services, you have to either spin up your own instance
of one or visit the site, then either create or search around for a test case.
This in and of itself is a problem because none of these tools other than
GitHub are our main code host that we use in our day to day. This means that we
aren't entirely used to the code host and creating or finding test cases will
likely be a cumbersome task. Our solution will eliminate this problem by
creating test cases for us. Each test case will be the same code for each code
host making it easy to navigate to certain scenarios we'd like to test. In
addition to this, creating or modifying test cases will be easy because we do
it in one place and they are generated for each code host automatically.

### Alternatives considered

One alternative considered was to spin up code hosts via docker containers,
manually get them into a "known state" where we can run unit tests against
them. This would be nice because it wouldn't rely on web scraping or other
"hacked together" solutions to get the code host into a known state. The
problem with this solution is that we'd either have to leave the containers
running somewhere in our infrastructure or save volumes of the data bases and
repository state somewhere so that we can restart the code host on any machine
and then point it to the volume so that it is in the same state. This is a
problem because it relies on some external state that comes with a maintinence
overhead we'd like to avoid. This solution is similar to our current situation
in that it relies on external services, which we have recently decided to
remove from our k8s cluster.

Another deviation from the plan that we've landed on was writing unique test
for each code host rather than spending the time to abstract the tests and
generate them for each code host. This may be quicker up front but will be
harder to maintain, particularly when we need to change something about a test
case. This will require us to change something n times rather than once. Our
current solution is beneficial because we generally want to test the same
scenario on each code host. If we want specific tests for a specific code host,
there is nothing stopping us from writing them on top of the abstracted tests.

## Checklist

- [ ] Implement code host functions for each code host
  - [ ] GitLab
    - [ ] Initialize (1 hr - via docker)
    - [ ] Add repository (2 hr)
    - [ ] Create code review (2 hr)
    - [ ] Update code review (1 hr)
    - [ ] Complete code review (1 hr)
    - [ ] View code review (0.5 hr)
    - [ ] View repo/rev/file (0.5 hr)
    - [ ] Get token at position for code view of type (1 hour - uses existing functions)
    - [ ] Shutdown (0.5 hr)
  - [ ] GitHub.com
    - [ ] Initialize (0 hr - uses github.com)
    - [ ] Add repository (1 hr)
    - [ ] Create code review (1 hr)
    - [ ] Update code review (0.5 hr)
    - [ ] Complete code review (0.5 hr)
    - [ ] View code review (0.5 hr)
    - [ ] View repo/rev/file (0.5 hr)
    - [ ] Get token at position for code view of type (1 hour)
    - [ ] Shutdown (0 hr)
  - [ ] GitHub Enterprise
    - [ ] Initialize (2 hr - use existing k8s infra)
    - [ ] Shutdown (0.5 hr)
    - [ ] Abstract GitHub.com implementation to not hard code GitHub.com into URLs (1 hr)
  - [ ] Phabricator
    - [ ] Initialize (1 hr - via docker)
    - [ ] Add repository (2 hr)
    - [ ] Create code review (2 hr)
    - [ ] Update code review (1 hr)
    - [ ] Complete code review (1 hr)
    - [ ] View code review (0.5 hr)
    - [ ] View repo/rev/file (1 hr)
    - [ ] Get token at position for code view of type (1 hr)
    - [ ] Shutdown (0.5 hr)
  - [ ] BitBucket Server
    - [ ] Initialize (2 hr - use existing k8s infra)
    - [ ] Add repository (2 hr)
    - [ ] Create code review (2 hr)
    - [ ] Update code review (1 hr)
    - [ ] Complete code review (1 hr)
    - [ ] View code review (2 hr)
    - [ ] View repo/rev/file (1 hr)
    - [ ] Get token at position for code view of type (1 hr)
    - [ ] Shutdown (0.5 hr)
- [ ] Abstract current E2E tests to work on each code host
  - [ ] Use code host functions to view specific files/code reviews rather than hard coded links (1 hr)
  - [ ] Modify tests to use not use hard coded GitHub class names
    - [ ] GitLab (1 hr)
    - [ ] BitBucket Server (1 hr)
    - [ ] Phabricator (1 hr)

## Done date

2/28/19

This is a rough estimate and is subject to change. It will change based on
changing scope around which code hosts make it into the v0 implementation of
this solution and how many tests that are implemented.

## Retrospective

[This section is completed after the project is completely done (i.e. the checklist is complete).]

### Actual checklist

[The checklist that was actually completed (i.e. paste the final checklist from the issue). Explain any differences from the original checklist in the plan.]

### Actual done date

[The date that the project was actually finished. Explain why this is earlier or later than originally planned or explain why the project was not completed.]
