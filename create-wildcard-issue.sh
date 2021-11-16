#!/usr/bin/env bash
COMPONENT_NAME=$1
COMPONENT_FRIENDLY_LOWERCASE_NAME=$2

# Tracking issue
TRACKING_ISSUE_TITLE="Wildcard V2: \`$COMPONENT_NAME\` Tracking issue"
TRACKING_ISSUE_BODY="
### Plan

<!--
Summarize what the team wants to achieve this iteration.
- What are the problems we want to solve or what information do we want to gather?
- Why is solving those problems or gathering that information important?
- How do we plan to solve those problems or gather that information?
-->

Tracking Issue for individual \`$COMPONENT_NAME\` component

### Tracked issues

<!-- BEGIN WORK -->
<!-- END WORK -->

#### Legend

- 👩 Customer issue
- 🐛 Bug
- 🧶 Technical debt
- 🎩 Quality of life
- 🛠️ [Roadmap](https://docs.google.com/document/d/1cBsE9801DcBF9chZyMnxRdolqM_1c2pPyGQz15QAvYI/edit#heading=h.5nwl5fv52ess)
- 🕵️ [Spike](https://en.wikipedia.org/wiki/Spike_(software_development))
- 🔒 Security issue
- 🙆 Stretch goal
- :shipit: Pull Request
"

gh issue create --title="$TRACKING_ISSUE_TITLE" --body="$TRACKING_ISSUE_BODY" --label="wildcard-v2" --label="wildcard-v2/new-components" --label="wildcard-v2/$COMPONENT_FRIENDLY_LOWERCASE_NAME" --label="team/frontend-platform" --label="tracking"

# Conditionally create Implementation issue
if [ "$IMPLEMENTATION_ESTIMATE" ]; then
  IMPLEMENTATION_ISSUE_TITLE="Wildcard V2: \`$COMPONENT_NAME\` Implementation"
  IMPLEMENTATION_ISSUE_BODY="
## Wildcard implementation of the \`$COMPONENT_NAME\` component.

See [Wildcard V2 - Planned work](https://docs.google.com/document/d/1NisbJPiadtt5jQw4vUUYJOr8dD6Urwn5V0mVq2wt6bE/edit#heading=h.g4cw92w3ouhw) for more context

## Acceptance criteria
- [ ] Component is implemented within Wildcard
- [ ] Component has tests
- [ ] Component is accessible
- [ ] Component matches [Figma designs](https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A1)
"

  gh issue create --title="$IMPLEMENTATION_ISSUE_TITLE" --body="$IMPLEMENTATION_ISSUE_BODY" --label="wildcard-v2" --label="wildcard-v2/new-components" --label="wildcard-v2/$COMPONENT_FRIENDLY_LOWERCASE_NAME" --label="team/frontend-platform" --label="estimate/$IMPLEMENTATION_ESTIMATE"
fi

# Conditionally create Codemod issue
if [ "$CODEMOD_ESTIMATE" ]; then
  CODEMOD_ISSUE_TITLE="Wildcard V2: \`$COMPONENT_NAME\` Codemod"
  CODEMOD_ISSUE_BODY="
## Codemod to migrate any existing usage of this pattern to use the \`$COMPONENT_NAME\` Wildcard component.

See [Wildcard V2 - Planned work](https://docs.google.com/document/d/1NisbJPiadtt5jQw4vUUYJOr8dD6Urwn5V0mVq2wt6bE/edit#heading=h.g4cw92w3ouhw) for more context

## Acceptance criteria
- [ ] Codemod created in https://github.com/sourcegraph/codemod
- [ ] Codemod has been ran on all relevant files in the Sourcegraph repository
"

  gh issue create --title="$CODEMOD_ISSUE_TITLE" --body="$CODEMOD_ISSUE_BODY" --label="wildcard-v2" --label="wildcard-v2/new-components" --label="wildcard-v2/$COMPONENT_FRIENDLY_LOWERCASE_NAME" --label="team/frontend-platform" --label="estimate/$CODEMOD_ESTIMATE"
fi

# Conditionally create Lint Rule issue
if [ "$LINT_RULE_ESTIMATE" ]; then
  LINT_RULE_ISSUE_TITLE="Wildcard V2: \`$COMPONENT_NAME\` Lint rule"
  LINT_RULE_ISSUE_BODY="
## Eslint rule to enforce the \`$COMPONENT_NAME\` Wildcard component.

See [Wildcard V2 - Planned work](https://docs.google.com/document/d/1NisbJPiadtt5jQw4vUUYJOr8dD6Urwn5V0mVq2wt6bE/edit#heading=h.g4cw92w3ouhw) for more context

## Acceptance criteria
- [ ] Eslint rule is added to enforce usage
- [ ] Eslint rule is implemented in the Sourcegraph repository
"

  gh issue create --title="$LINT_RULE_ISSUE_TITLE" --body="$LINT_RULE_ISSUE_BODY" --label="wildcard-v2" --label="wildcard-v2/new-components" --label="wildcard-v2/$COMPONENT_FRIENDLY_LOWERCASE_NAME" --label="team/frontend-platform" --label="estimate/$LINT_RULE_ESTIMATE"
fi

# Conditionally create manual migration issue
if [ "$MANUAL_MIGRATION_ESTIMATE" ]; then
  MANUAL_MIGRATION_ISSUE_TITLE="Wildcard V2: \`$COMPONENT_NAME\` Manual migration"
  MANUAL_MIGRATION_ISSUE_BODY="
## Manual migration required to update our code to use the \`$COMPONENT_NAME\` Wildcard component.

See [Wildcard V2 - Planned work](https://docs.google.com/document/d/1NisbJPiadtt5jQw4vUUYJOr8dD6Urwn5V0mVq2wt6bE/edit#heading=h.g4cw92w3ouhw) for more context

Note: Before doing this, we should check if there is scope for a codemod to automate this migration. If so, the codemod should be implemented and ran before starting work on this issue.

## Acceptance criteria
- [ ] All incorrect usage of the previous pattern has been updated to use the Wildcard component
"

  gh issue create --title="$MANUAL_MIGRATION_ISSUE_TITLE" --body="$MANUAL_MIGRATION_ISSUE_BODY" --label="wildcard-v2" --label="wildcard-v2/new-components" --label="wildcard-v2/$COMPONENT_FRIENDLY_LOWERCASE_NAME" --label="team/frontend-platform" --label="estimate/$MANUAL_MIGRATION_ESTIMATE"
fi
