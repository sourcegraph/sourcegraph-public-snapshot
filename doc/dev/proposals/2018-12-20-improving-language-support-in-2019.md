# Improving language support in 2019

## Background

Improving language support contributes to the Q1 2019 goal of increasing the number of Sourcegraph instances that go from 1 to 20 users because: 

- many admins want to see their language(s) supported before sharing Sourcegraph with their team; and
- good language support is necessary for the code host integration (via the browser extension) to be useful, and that's easy to spread.

Despite deploying a ton of experimental language servers in 2018, language support has not improved much in the past year - why?

- Wrapping language servers gave us remote execution of the language server, but lsp-adapter did not solve more fundamental problems related to incompatible compiler and package manager versions, not installing dependencies, slow initialization, and poor quality in general
- lsp-proxy was complex and no single Sourcegrapher fully understood how lsp-proxy, lsp-adapter, and indexer worked
- Aside from Swift, we simply didn’t prioritize work on language servers (partly because of the complexity of lsp-proxy, partly because of other priorities)

Many things have changed in the last year to make it possible to improve language support now:

- With [Sourcegraph extensions](https://docs.sourcegraph.com/extensions), it’s easier to understand how code intelligence on Sourcegraph works, which makes it easier to build on and all you need to understand is the Sourcegraph extension API (no need to understand [xfiles/xcontents](https://github.com/sourcegraph/language-server-protocol/blob/master/extension-files.md), [xcache](https://github.com/sourcegraph/language-server-protocol/blob/master/extension-cache.md), [lsp-adapter](https://github.com/sourcegraph/lsp-adapter), lsp-proxy, etc.) in order to add language support
- There’s a new [Swift language server (apple/sourcekit-lsp)](https://github.com/apple/sourcekit-lsp)
- There’s a new [Python language server (Microsoft/python-language-server)](https://github.com/Microsoft/python-language-server)
- We’ve learned that it’s fairly easy to patch existing language servers (Go, TypeScript, and Python) to support zip archive fetching and WebSockets. This results in a more maintainable and "pure" language server than wrapping a language server with lsp-adapter.
- We’ve learned that shipping experimental language servers is not an effective way to attract community/contractor help or useful feedback. (We [deactivated experimental language servers](https://about.sourcegraph.com/blog/java-php-experimental-language-servers-temporarily-unavailable).)

Based on these learnings, the following principles will guide future improvements:

- Prioritize languages by a combination of popularity and ease of analysis (statically-typed languages, ones that already have solid language servers, ones that customers are asking for, etc.)
- Focus on quality over quantity (already tried quantity in 2018 with lsp-adapter and cold emails, and this stagnated)
- Pay contractors if it saves us time and effort (they’re experts in the respective language and will likely be able to hit the ground running)
- UI/UX ergonomics matter: suppress non-actionable errors and indicate when some analysis is taking a long time

## Proposal

Put effort towards one language at a time in rough order of prospective deals for Sourcegraph:

- ✅ JavaScript/TypeScript using https://github.com/sourcegraph/lang-typescript
- 📝 Java using https://github.com/beyang/eclipse.jdt.ls/tree/wip
- 📝 Python using https://github.com/sourcegraph/sourcegraph-python
- ✅ Go using https://github.com/sourcegraph/lang-go
- 📝 Swift using https://github.com/apple/sourcekit-lsp

For each language, patch the language server as needed and build a Sourcegraph extension that communicates with it and satisfies the test plan below.

### Test plan

Manually open each link to verify that the language server is working. In rough order of increasing difficulty:

- [ ] Link to ANY working {hover,definition,references}
- [ ] Link to 3 working {hover,definition,references}, each in a different popular repo (e.g. for Go https://github.com/search?q=stars%3A%3E0+language%3Ago&type=Repositories)
- [ ] Initialization time on popular repos is <10s
- [ ] After initialization on popular repos, most response times are <3s
- [ ] Monorepo support (initialization time and response time scale sublinearly with repo size)
- [ ] Cross-repo code intelligence on popular repos

### Release plan

For each language, publish a Sourcegraph extension with usage instructions in their READMEs and then iterate and address feedback until customers actually use the language extension on a regular basis.

## Rationale

[A discussion of alternate approaches and the trade offs, advantages, and disadvantages (including risks and uncertainties) of the specified approach.]

## Checklist

- [ ] By Dec 21, publish a modified version of this proposal that communicates our high level plan and states what Sourcegraph wants in terms of language support. Briefly summarize and link to the master plan. Mention that we'll help with sponsorship, technical advice, getting feedback, etc.
- [ ] By Dec 21, update langserver.org and our careers repo to link to this blog post and have @sqs to tweet this to Future of Coding
- [ ] By Dec 21, after doing an initial triage, publish one blog post per language similar to the above blog post but with more details about the particular language.
- [ ] By Jan 2, add an Implementation section to the README of lang-{go,typescript,python}
- [ ] By Jan 2, prepare a doc outlining how to build language extensions building off of https://github.com/sourcegraph/sourcegraph/pull/628 but avoid recommending certain approaches (leave it up to the language server author), link to the Implementation sections above
- [ ] By Jan 4, basic Python support is deployed to Sourcegraph.com (@sqs has already implemented most of this)
- [ ] By Jan 11, Python has cross-repo support
- [ ] By Jan 25, basic Swift support is deployed on Sourcegraph.com
- [ ] By Feb 1, Swift has cross-repo support
- (these dates are highly tentative because they depend on external factors such as the quality of the new Swift language server and how hard it might be to find contractor(s))

## Done date

Tentatively 1-3 months per language throughout the year.

## Retrospective

[This section is completed after the project is completely done (i.e. the checklist is complete).]

### Actual checklist

[What is the actual checklist the you completed (i.e. paste the final checklist from the issue here)? Explain any differences from the original checklist in the proposal.]

### Actual done date

[What is the date that the project was actually finished? Explain why this is earlier or later than originally planned or explain why the project was not completed.]
