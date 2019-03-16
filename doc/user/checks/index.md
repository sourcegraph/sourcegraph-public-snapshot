# Checks

Checks let you define rules that are continously applied to your code, with a developer-friendly workflow. It's intended for team process and rules that are too complex to implement in a linter or CI, such as:

- Don't commit API keys
- All code must have an owner
- 24-hour SLA for code reviewers to review changes assigned to them
- Notify authors who have simultaneous unmerged, conflicting changes
- New dependencies require tech lead approval
- Require review for changes to designated sensitive areas of the code *and their transitive dependencies*
- Assign crash reports to the author of the most recent changed line in the stack trace
- Use consistent branding and language in product text strings

Every team follows some rules like these, but the process is manual and informal (or relies on a clunky internal tool). Checks let you define these processes in code (**PaC**, Process as Code) so they're automatic and explicit.

**For developers:** Checks let your team automate away many of the things that delay and frustrate you, so you can focus on coding and get your code merged sooner. Automated testing, [CI](https://en.wikipedia.org/wiki/Continuous_integration), and [CD](https://en.wikipedia.org/wiki/Continuous_deployment) have made great progress eliminating delays and frustration, and this is the next logical step.

**For enginering managers:** You spend a ton of time helping your team adhere to these manual and informal processes---and dealing with the fallout from avoidable mistakes. Checks give you confidence that the processes are being followed. This means you deal with only the exceptional cases that merit your attention, and you have much more time to spend on higher-value tasks.

<!-- **For internal tools teams:** Checks are a core part of Sourcegraph, which handles all the complexity of authenticating, syncing, storing, scaling, and serving the repository and other data. TODO, focus more on the problems it solves for internal tools teams instead of assuming they want to build it in house and arguing against it -->

These rules and processes are different from CI and lint because they: <!-- TODO: rough -->

- Depend on information defined outside of the repository
- Depend on information that can change at any time, not just with a commit
- Are gradually enforced
- Consistency across all repositories is needed (e.g., if a repository forgets to add the right CI config to check for critical security issues, that could lead to those being missed if you are relying on CI)
- Need custom approval workflows, such as a new dependency being added to a global allowlist by the tech lead

## Use cases

### Don't commit API keys



<!--

Sourcegraph extensions lets your team all use the same tools. Sourcegraph checks lets your team all use the same rules.

Checks is like having an extra tech lead, security engineer, and architect on each team. Price it like the cost of those folks' salaries?

PaC: Process as Code (like IaC, infrastructure as code)

-->
