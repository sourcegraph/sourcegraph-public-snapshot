# Why code search is still needed for monorepos

Developers who work in monorepos sometimes think code search (such as [Sourcegraph](https://about.sourcegraph.com)) isn't useful for them:

> "We use a monorepo, so we don't need a code search tool. I can search the entire monorepo in my editor or using `ripgrep`/`grep`. My editor gives me go-to-definition and find-references across the entire monorepo."

Many developers told us they initially felt this way, but they later came to *love* using a code search tool on their monorepo. [We asked them](https://twitter.com/sqs/status/1325643096588230658) to find out what arguments would have convinced them earlier.

Here are the best arguments for using a code search tool on a monorepo, from monorepo-using devs who were initially skeptical but changed their mind and now love code search:

1. [You can share links to a code search tool](#share-links)
1. [Using a separate code search tool helps (not hurts) your flow](#flow)
1. [Google, Facebook, and Twitter have monorepos and use code search heavily](#social-proof)

## You can share links to a code search tool {#share-links}


**Scenario:** You're deep in flow, but then a team member asks you a question about how some other code works, or something like that.

- 🤬 Using your editor's search, you can find the answer. But then how do you share that with your team member? Screenshot your editor? Type out "see lines 38-40 of `client/web/src/user/account/AccountForm.tsx` …"?
- 😊 With a code search tool, you'd just copy the URL and paste it to the other person. They can visit the link and see what you mean instantly.

**Another scenario:** You're coding and come across a bug that, upon a quick search in your editor, is present in dozens of files (such as a typo, function misuse, etc.).

- 🤬 You post an issue, but how do you link to all the instances of this bug? You can't link to your editor's search. You could describe it ("we need to fix everywhere where `updateAccount` is used in the client code …"), but that's imprecise and hard to understand.
- 😊 With a code search tool, you can just link to a search results page with all of the places that need fixing.

## Using a separate code search tool helps (not hurts) your flow {#flow}

When you're coding and need to search or navigate to answer some question, it's really nice to stay in flow and be able to jump right back into writing code when you get the answer.

It might *seem* like staying in a single tool (your editor) means staying in flow, but that's not true. Here are some examples to illustrate the point.

**Scenario:** You're writing a call to a function and want to see usage examples or patterns from elsewhere in your codebase. Your editor's find-references panel works for simple cases.


- 🤬 But if you want to jump to other call sites and poke around the code for each in your editor, you've opened a ton of new tabs and *lost your editing flow*.
- 😊 A separate code search tool would let you preserve your blinking cursor and editor state so you can jump right back in when you find the answer.

    > "Preserving my editing flow is the primary reason why I use browser-based code search tools for looking up things."
- 😊 You can also keep the code search tool up in a separate window (or monitor) to easily refer back to the usage examples/patterns while coding. (Editors get weird with multiple windows.)
- 😊 If you need to filter a long list of matches or references by subdirectory, arguments, or something else, that's usually cumbersome or impossible in your editor (but easy in code search).

**Another scenario:** You're deep in flow on your branch, but then you get a bug report and need to triage it.

- 🤬 You need to stash your changes and check out the main branch, then search and navigate the code in your editor. Your dev server and test watcher get messed up, and you lose your editing flow. Your editor locks up as it starts reindexing/reanalyzing a different branch.
- 😊 A separate code search tool would let you quickly triage a bug on any branch without changing your local branch or affecting your dev setup.
- 😊 The code search tool can show much more helpful code context, such as [Git blame information after each line](https://sourcegraph.com/extensions/sourcegraph/git-extras), [code coverage overlays](https://sourcegraph.com/extensions/sourcegraph/codecov), runtime info from [Datadog](https://sourcegraph.com/extensions/sourcegraph/datadog-metrics)/[LightStep](https://sourcegraph.com/extensions/sourcegraph/lightstep)/[Sentry](https://sourcegraph.com/extensions/sourcegraph/sentry)/etc., static analysis and lint results from [SonarQube](https://sourcegraph.com/extensions/sourcegraph/sonarqube), and more. You could configure some of these things to display in your editor, but that's cumbersome and they're noisy for the majority of the time when you're writing code.
- 😊 A code search tool does all the hard work (indexing and analysis) on the server beforehand, so your local machine remains fast and responsive.
 
    > "The JetBrains IDEs have great search capabilities. However, indexing a large repo is slow and draining on even a powerful MacBook Pro, and that happens every time you switch to another branch."
- 😊 If you identify a likely culprit (such as a problematic line of code) via code search, it's easy to get a permalink to that line to add to the bug report.


## Google, Facebook, and Twitter have monorepos and use code search heavily {#social-proof}

These 3 companies are known for using monorepos internally. Their devs all heavily rely on code search when working in their monorepos:

- Google has a [monorepo](https://research.google/pubs/pub43835/) and reports that [Google devs use code search 12 times per workday on average](https://research.google/pubs/pub43835/).
- Facebook has a [monorepo](https://www.facebook.com/atscaleevents/videos/systems-scale-2019-monorepos-moving-fast-in-a-huge-repository/457153524992062/) and [code search is heavily used by devs for everyday questions and to enable refactoring and analysis](https://www.facebook.com/atscaleevents/videos/1911812842425144/).
- Twitter has a [monorepo](https://www.youtube.com/watch?v=IL6LBWNi3fE) and [uses code search](https://twitter.com/willnorris/status/1311043937784590336).

These monorepos are large, of course. If you have a small monorepo, then you should probably disregard this argument (but heed the others).

---

## Feedback

Do you have additional or better arguments? Do you disagree with any of these arguments? Use the **Edit this page** link or contact us ([@srcgraph](https://twitter.com/srcgraph)) to suggest improvements.

Disclaimer: [Sourcegraph](https://about.sourcegraph.com) is a universal code search company, so of course we would say these things, right? But our intrinsic love for code search came first. It's *why* we started/joined Sourcegraph and want to [bring code search to every dev](https://about.sourcegraph.com/company/strategy).
