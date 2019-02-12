# 3.0 beta retrospective

[Survey](https://docs.google.com/forms/d/e/1FAIpQLSceowJsPCfow-7cwSiQk4xpQkwu_az6sQ2xIWbSLGEbnLePMQ/viewform?usp=sf_link)

## Action items

1. Nick, Quinn, Beyang [document product/engineering process](https://github.com/sourcegraph/sourcegraph/pull/1961)
    - Communicate that we hold our release date fixed and adjust scope (or apply beta labels) as necessary.
    - Include owners for design, code review, and testing in plans.
    - Nick to review owners for each project.
2. Chris lead discussion about enabling basic code intel extension by default.
3. Tomás propose a plan for [safe update process](https://github.com/sourcegraph/sourcegraph/pull/2014)

## Start

* +7 Shorter release cycles (1)
   * +1 Be more allergic to visual noise (non-actionable errors, notifications, etc.)
* +5 The tech lead should make sure that tasks are assigned to the right person / paired with the right people (i.e. - don't have someone inexperienced in a certain project area work on a new feature that we have a hard deadline for) (1)
   * Example: refactor in area that wasn't familiar with, someone else on the team had a lot more context on that area
* +4 Account for code review, testing, and design in the planning process for projects (1)
   * Test + review “buddies”
* +4 Better unit, integration and end to end testing. 
   * Manual testing: Test earlier during development, maybe by designating testers for each project early on. Can go together with code review. This would reduce time spent on testing and prevent discovering many/big issues the week of the release. (1)
   * Geoffrey, Tomás, Beyang
* +3 Make setup easier for customers (2, 3)
   * Be aware of how little effort most people are willing to put into setting up an instance
   * Find a way to make Sourcegraph's upgrade process as painless and automatic as possible (one button click for most deployments), thus allowing us to reap the benefits of continuous delivery.
   * Keegan, Chris, Geoffrey, Beyang, Tomás
* +2 Write down plans for projects that capture context, goals, and decisions (1)
   * Get a firm grasp on the technical approach to implementing a feature before starting
* +2 Triage all issues in the milestone at the beginning of each release cycle (1)

## Stop

* +4 If an issue is removed from the current milestone, automatically just pushing it forward to the next milestone
* +4 Trading off as much quality for speed.
* +4 Crunching
* +4 it'd be nice to know ASAP whenever we have a hard goal that we need to hit so that we can plan as effectively as possible. (1-20DAU goal, +3 companies in Jan)

## Continue

* +3 Do not postpone tackling the technical debt introduced by a rushed release. (how the time between 3.0 beta and 3.0 is entirely tech debt and bugs, not features)
* +1 Ensure consistency in e.g. language extension settings
   * ties into planning, holistic testing sooner
   * example: renaming site config settings causes hassle
   * backcompat vs. consistency
