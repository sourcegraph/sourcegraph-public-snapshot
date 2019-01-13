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

- Prioritize languages by a combination of popularity, ease of analysis, and Sourcegraph customer needs (statically-typed languages, ones that already have solid language servers, ones that customers are asking for, etc.)
- Focus on quality over quantity (already tried quantity in 2018 with lsp-adapter and cold emails, and this stagnated)
- Pay contractors if it saves us time and effort (they’re experts in the respective language and will likely be able to hit the ground running)
- UI/UX ergonomics matter: suppress non-actionable errors and indicate when some analysis is taking a long time

## Proposal

Put effort towards one language at a time:

- ✅ JavaScript/TypeScript using https://github.com/sourcegraph/lang-typescript
- 📝 Java using https://github.com/beyang/eclipse.jdt.ls/tree/wip
- 📝 Python using https://github.com/sourcegraph/lang-python
- ✅ Go using https://github.com/sourcegraph/lang-go
- 📝 Swift using https://github.com/apple/sourcekit-lsp

For each language, patch the language server as needed and build a Sourcegraph extension that communicates with it and satisfies the test plan below.

### Test plan

When developing support for a language, copy this template below and customize it for the particular language you're working on.

Manually open each link to verify that the language server is working. In rough order of increasing difficulty:

- [ ] Link to ANY working {hover,definition,references}
- [ ] Link to 3 working {hover,definition,references}, each in a different popular repo (e.g. for Go https://github.com/search?q=stars%3A%3E0+language%3Ago&type=Repositories)
- [ ] Initialization time on popular repos is <10s
- [ ] After initialization on popular repos, most response times are <3s
- [ ] Monorepo support (initialization time and response time scale sublinearly with repo size)
- [ ] Cross-repo code intelligence on popular repos

@chrismwendt and @felixfbecker will review the repository and performance benchmark choices and modify them in order to accurately represent common usage on Sourcegraph.

### Release plan

For each language:

- Publish a Sourcegraph extension with usage instructions in their READMEs
- Iterate and address feedback until customers actually use the language extension on a regular basis

## Rationale

One alternative is for Sourcegraphers to try to implement everything. That would probably be be inefficient considering our lack of expertise in some languages.

Risks and unknowns:

- Communication between browser <-> language server: most language servers speak STDIO only (possible solution: support WebSockets or HTTP)
- Obtaining files: most language servers expect the repository to already exist somewhere on disk (possible solution: fetch zip URL)
- Arbitrary code execution in package managers (possible solutions: disable running scripts in the package manager like yarn’s --ignore-scripts, run in gvisor/VM)
- Slow initialization: many language servers build the entire project before responding to requests (possible solutions: cache build results, lazily fetch dependencies, lazily build the AST, lazy everything)
- Inability to build the project due to varying (i.e. commit-specific and/or ambiguous) versions of compiler, package manager, etc.
- Language server authors might not accept the changes we make (WebSocket support, zip archive fetching, etc.)

## Checklist

- [ ] By Dec 21, publish a blog post with modified version of this proposal that communicates our high level plan and states what Sourcegraph wants in terms of language support. Briefly summarize and link to the master plan. Mention that we'll help with sponsorship, technical advice, getting feedback, etc.
- [ ] By Dec 21, update langserver.org and our careers repo to link to this blog post and have @sqs to tweet this to Future of Coding
- [ ] By Dec 21, after doing an initial triage, publish one blog post per language similar to the above blog post but with more details about the particular language.
- [ ] By Jan 2, add an Implementation section to the README of lang-{go,typescript,python}
- [ ] By Jan 2, prepare a doc outlining how to build language extensions building off of https://github.com/sourcegraph/sourcegraph/pull/628 but avoid recommending certain approaches (leave it up to the language server author), link to the Implementation sections above
- [ ] By Jan 2, update langserver.org entries with check boxes for any new and relevant pieces of functionality based on the Implementation sections above
- [ ] By Jan 4, basic Python support is deployed to Sourcegraph.com (@sqs has already implemented most of this)
- [ ] By Jan 9, same-file Ruby hovers work on GitHub.com
- [ ] By Jan 11, Ruby hovers work for full projects
- [ ] By Jan 14, Python has cross-repo support
- [ ] By Jan 18, determine quality of Swift language server
- [ ] By Jan 21, find a Swift contractor
- [ ] By Jan 25, basic Swift support is deployed on Sourcegraph.com
- [ ] By Feb 1, Swift has cross-repo support

## Done date

Python, Swift, and (some) Ruby support will be done by Feb 1.

## Retrospective

[This section is completed after the project is completely done (i.e. the checklist is complete).]

### Actual checklist

[What is the actual checklist the you completed (i.e. paste the final checklist from the issue here)? Explain any differences from the original checklist in the proposal.]

### Actual done date

[What is the date that the project was actually finished? Explain why this is earlier or later than originally planned or explain why the project was not completed.]
