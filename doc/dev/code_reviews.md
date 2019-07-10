# Code reviews

All code should be reviewed and approved by an appropriate teammate before being merged into the `master` branch.

## Why do we require peer code reviews?

Code reviews benefit everyone on the team, and the team itself.

For authors:

- Requesting a code review from one of your peers is motivation to ensure that the quality of your code is high.
- Writing a good PR description (and commit messages) develops your technical communication skills.
- Receiving code review feedback can give you valuable insight into aspects of your change that you hadn't considered (architectural, performance, etc).
- Receiving feedback about how your code could be improved helps you learn how to write better code in the future.

For reviewers:

- Reviewing code from others increases your code literacy and empathy. You will learn new things from how your peers write code and you will be more aware of how to write code that is easier for your reviewers to understand
- Reading PR descriptions increases your technical communication literacy and empathy. You will learn what makes a PR description effective and you will be more aware of how to write of how to write PR descriptions that are easier for your reviewers to understand.
- Being a reviewer gives you the ability to share knowledge you have that others do not. Every person, regardless of experience level, has both something to learn and something to contribute.

For the team:

- Code reviews increase the overall quality of our codebase by catching design and implementation defects before changes ship to customers.
- Code reviews increase the overall consistency of our codebase by spreading knowledge of best practices and eliminating anti-patterns.
- Code reviews distribute domain expertise, code knowledge, and ownership across the team so development can continue when people are unavailable (e.g. vacation).

## When is a code review not required?

We do not technically prevent PRs from being merged before being explicitly approved because there exist PRs that are appropriate to merge without waiting for a review.

Some examples:

- Reverting a previous change to solve a production issue.
- Minor docs changes.
- Auto-generated changes, such as when performing a version release, where a reviewer would not have anything of substance to review.

It should be obvious to any other engineer on the team from the PR description and/or diff that it was appropriate for you to not wait for approval.

Here are some examples of reasons to skip code review that are NOT acceptable:

- "I promised that I would ship this to $CUSTOMER by $DEADLINE"
    - The customer expects the feature to work and be maintained. Code review helps ensure both of these things by increasing the quality and distributing ownership.
- "This code is experimental"
    - If the experiment is a success and the code is useful, then it will be easier to have done code reviews incrementally than to expect someone to review the code after the fact. The latter rarely happens successfully.
- "I don't have someone to review this code"
    - You should find a new person to distribute knowledge, expertise, and context to.

## What makes an effective code review?

[This blog post](https://medium.com/@schrockn/on-code-reviews-b1c7c94d868c) effectively explains some principals which our team values.

The code author and code reviewer have a _shared goal_ to bring the PR into a state that meets the needs that motivated the change and meets our quality standards.


Do:

- Take time to understand the context and goal of the PR. If it isn't clear, ask for clarification.
- Acknowledge what the author did well to balance the tone of the review.
- Prioritize reviewing the overall design and structure of the code over smaller code syle issues. Only comment on the latter after you explictly approve of the former.
- Make it clear which comments are blocking your explicit approval (e.g. use "nit:" prefix for minor comments) and approve if all of your comments are minor.
- If the author were to address all of your comments faithfully and you would be content, then you should also approve to avoid the author needing to wait for a subsequent review without reason (exception: you asked for fundamental or vast/large changes and believe those will need re-review by you).
- When you are making comments on a PR, use a tone that is kind, empathetic, collaborative, and humble. [Further reading](https://mtlynch.io/human-code-reviews-1/).
    | 😕 | 🤗 | Why? |
    |---|---|---|
    | You misspelled "successfully" | sucessfully -> successfully | Avoid using "you" because it can make the comment feel like a personal attack instead of just a minor correction. |
    | Rename this variable to `betterName`. | Can we rename this variable to be more descriptive? What do you think of `betterName`? | Feedback that is framed as a request encourages collaborative conversation. Including your reason helps the author understand why and potentially think of even better alternatives. |
    | Can we simplify this with a list comprehension? | Consider simplifying with a list comprehension like this: $EXAMPLE | Provide examples when suggesting changes so the author doesn't have to infer what you mean (which might be difficult if you refer to constructs that they haven't used before). |
    | Why didn't you $X? | What do you think about doing $X? I think it would help with $Y. | A "What" question is better than a "Why" question because the latter sounds accusatory. Either the author considered $X and decided not to, or they didn't consider it, but in either case it is frustrating to the author if the reviewer doesn't explain why they think $X should be considered. |
    | This code is confusing. Can you clarify it? | It was hard for me to understand $X because $Y, but I haven't been able to think of a way to improve it. Do you have any ideas? | Don't declare your opinion as fact; instead, share your subjective experience and try to offer a suggestion for improvement. If you don't have a suggestion, acknowledge that to show that you empathize with the difficulty of improving this code. | 

Don't

- Take longer than one business day to respond to respond to a PR that is ready for your review (or re-review).
- Have protracted discussions in PR comments. If it can't be settled quickly in a few review round trips, try discussing in person or on a video call because these mediums are higher bandwidth and encourage empathy. Then summarize the results in the PR discussion after the fact.

## Who should I get a code review from?

You should get a code review from the person who's approval will give you the most confidence that your change is high quality. If you are modifying existing code, you can use `git blame` to identify the previous author and the previous reviewer because they probably will have helpful context.

If your change touches multiple parts of our codebase (e.g. Go, TypeScript), then you might need to get approvals from multiple peers (e.g. a Go reviewers and a TypeScript reviewer).

GitHub will automatically assign reviewers if there is a matching entry in the [CODEOWNERS](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/.github/CODEOWNERS) file, but that doesn't necessarily mean that you need to wait for an approval from everyone. For example, if you are making a change to the search backend then you only need approval from one person on that team, not all of them.

Use your best judgement and ask if you are uncertain.
