# Code reviews

All code should be reviewed and approved by an appropriate teammate before being merged into the `master` branch.

## Why do we require peer code reviews?

Code reviews benefit everyone on the team, and the team itself.

For authors:

- Requesting a code review from one of your peers is motivation to ensure that the quality of your code is high.
- Writing a good PR description develops your technical communication skills.
- Having your code reviewed helps you learn how to write better code in the future because you will receive feedback about how your code could be improved.

For reviewers:

- Reviewing code from others increases your code literacy and empathy. You will learn new things from how your peers write code and you will be more aware of how to write code that is easier for your reviewers to understand
- Reading PR descriptions increases your technical communication literacy and empathy. You will learn what makes a PR description effective and you will be more aware of how to write of how to write PR descriptions that are easier for your reviewers to understand.

For the team:

- Code reviews increase the overall quality of our codebase by catching design and implementation defects before changes ship to customers.
- Code reviews increase the overall consistency of our codebase by spreading knowledge of best practices and eliminating anti-patterns.
- Code reviews distribute domain expertise, code knowledge, and ownership across the team so development can continue when people are unavailable (e.g. vacation).

## When is a code review not required?

We do not technically prevent PRs from being merged before being explicitly approved because there exist PRs that are appropriate to merge without waiting for a review.

Some examples:

- Reverting a previous change to solve a production issue.
- Minor docs changes.

It should be obvious to any other engineer on the team from the PR description and/or diff that it was appropriate for you to not wait for approval.

Here are some examples of unacceptable reasons to skip code review:

- "I promised that I would ship this to $CUSTOMER by $DEADLINE"
    - The customer expects the feature to work and be maintained. Code review helps ensure both of these things by increasing the quality and distributing ownership.
- "This code is experimental"
    - If the experiment is a success and the code is useful, it will have been easier to have done code reviews incrementally than expect someone to review the code after the fact.
- "I don't have someone to review this code"
    - It sounds like it is time to find a new person to distribute knowledge, expertise, and context to.

## What makes an effective code review?

Read https://medium.com/@schrockn/on-code-reviews-b1c7c94d868c; it is spot on.

As a reviewer, your job is to help the code author make their code better. The helpfullness of your code review is measured at their ear, not your mouth. 

Do:

- Take time to understand the context and goal of the PR. If it isn't clear, ask for clarification.
- Acknowledge what the author did well.
- Prioritize reviewing the overall design and structure of the code over smaller code syle issues. Only comment on the latter if you explictly approve of the former.
- Make it clear which comments are blocking your explicit approval (e.g. use "nit:" prefix for minor comments) and approve if all of your comments are minor.

Don't

- Have protracted discussions in PR comments. If it can't be settled quickly in a few review round trips, it is probably worth discussing in person or on a video call because these mediums are higher bandwidth and encourage empathy. Summarize the results in the PR after the fact.

