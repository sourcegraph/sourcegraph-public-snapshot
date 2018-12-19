# Extension-based Java code intelligence

## Background

We need to add back Java support via Sourcegraph extension. While doing so, we'll address some of
the persistent issues with old Java support:

- Fault-tolerant compilation: the [old Java language
  server](https://github.com/sourcegraph/java-langserver) was based on an internal `javac` compiler
  API. This API was not designed for the editor use case and consequently was not
  fault-tolerant. Considered alone, this would not be a serious issue as most code pushed to a
  source repository has a green build. However, the complexity of Java build systems (Gradle, Maven
  code generation plugins) meant that in a significant proportion of cases, we cannot construct the
  proper build environment in an automated fashion. In such cases, fault tolerance at the compiler
  level would mean we can still provide useful code intelligence even without a perfect build.
- Build configuration: Our support for Java build systems (Maven and Gradle) was incomplete and
  likely will always be incomplete due to the diversity and complexity of these build systems. We
  offered a fallback option, `javaconfig.json`, but this was a format specific to Sourcegraph and
  unlikely to be adopted as a standard in the Java community.
- Lack of community support: The old language server was closed source for most of its life. It
  seems unlikely that it will become the standard Java language server given its reliance on `javac`
  hampers its viability in editor plugins.
- Keeping up with new Java developments: Given its lack of community support, it would be up to us
  to keep the old language server up to date with new Java features. In the future, this will
  require a lot of ongoing work from Sourcegraph just to keep up to date with the latest in the Java
  community.

## Proposal

Implement Java code intelligence (feature parity with Java code intelligence in 2.13) via
[eclipse.jdt.ls](https://github.com/eclipse/eclipse.jdt.ls) and a corresponding Sourcegraph
extension.

### Test plan

* Add any appropriate unit tests to eclipse.jdt.ls
* Adapt the old integration test suite from the [legacy language server](https://github.com/sourcegraph/java-langserver)

### Release plan

* Update documentaton for Java language support
* Upload Java extension to Sourcegraph extension registry
* Docker image for running `eclipse.jdt.ls` in a way that's compatible with the Sourcegraph extension

## Rationale

Alternative 1: Adapt `github.com/sourcegraph/java-langserver` to a Sourcegraph extension. This is
not a good option due to the issues with `java-langserver`, mentioned in *Background*.

Alternative 2: Rely on basic code intelligence for Java indefinitely. Basic code intelligence falls
short of our customers' needs for accurate cross-repo code intelligence in the UI and as a basis for
automated tools in the future.

Alternative 3: Do not modify `eclipse.jdt.ls` directly, but use an intermediary service similar to
[`lsp-adapter`](https://github.com/sourcegraph/lsp-adapter). The intermediary would handle
LSP-websocket requests from the Sourcegraph extension, copy the repository from the Sourcegraph raw
API to local disk, and manage instances of `eclipse.jdt.ls` that operate on the local files it
fetches from the Sourcegraph raw API. We reject this option for the following reasons:

1. Operation complexity: The addition of an extra services makes this operationally more complex and
   harder to debug.
2. Performance: This architecture constrains us to always fetch the entire repo zip upfront, which
   we may not want to do for performance reasons. All of our existing language servers do this (with
   the exception of the legacy Java language server), but we would like the option to do on-the-fly
   file fetching in the future. Also, in the past, the languages supported by `lsp-adapter` were not
   very reliable or performant.
3. Future extensibility: We would like to build institutional knowledge of how the Java language
   server works, so we can modify and extend it in response to customer asks in the future. Doing it
   this way effectively treats `eclipse.jdt.ls` as a black box.

Note: I am not entirely sold on the above arguments. In particular, I think the primary bottlenecks
for language servers will be dependency resolution/fetching, not source file fetching. Furthermore,
if on-the-fly file fetching does become a bottleneck (and it probably will for monorepos), the
natural solution to me seems to be to fetch repositories by *build unit*, rather than file-by-file
fetching over a full VFS interface. The additional operational complexity I don't think is
significant (and will be outweighed by the added complexity of converting language servers to use a
VFS interface). Institutional knowledge we can develop over time in either situation. Currently, all
our existing language servers (with the exception of `java-langserver`, which we're abandoning)
fetch the entire repo zip upfront
([discussion](https://sourcegraph.slack.com/archives/CCLF4R6EM/p1544660870180700)).

However, given what we know now, I think it is reasonable to rule out this alternative for the time
being and go with the proposed path (modifying `eclipse.jdt.ls` to support websockets and remote
file fetching natively).  We can always revisit this alternative down the road. The signals that we
should revisit will be if we find we're reimplementing the following things in the same manner for
each language server:

* Upfront file fetching, either entire repo upfront or entire subtrees at once
* Translation from remote to local file URIs in every LSP handler (i.e., not treating remote file
  URIs "natively")
* Websockets

## Implementation


- [ ] By Jan 1: deploy alpha Java extension to sourcegraph.com. It supports a minimal set of Java
      repositories (e.g., Eclipse-based projects, vanilla Maven)
  - [x] Websocket support
  - [x] Remote file fetching support
  - [ ] Work through bugs that prevent the LS from supporting all operations it supports locally via
        the VS Code Java plugin
  - [ ] Test on large repositories to get a sense of performance at scale
- [ ] By Jan 7: Cross-repository j2d/find-refs for repos with supported build configurations
- [ ] By Jan 14: Maven support parity with old `java-langserver`, as measured by old integration
      test suite
- [ ] By Jan 18: Gradle support parity with old `java-langserver`
- [ ] By Feb 8 (3.1 release date):
  - [ ] Extension and language server reviewed and committed to `master`
  - [ ] Language server deployed to dogfood.
  - [ ] Language server deployed to sourcegraph.com.
  - [ ] Extension published to sourcegraph.com
- [ ] Post Feb 8: Customer and OSS followthrough
  - [ ] Walk major Java customers through upgrade process
  - [ ] Deal with any regressions/issues upgrading from 2.x to 3.1. Budget 2 weeks of time over the
        next month to deal with these.
  - [ ] Upstream changes to `eclipse.jdt.ls`

### Known risks and mitigations

- Major regressions at customers from `java-langserver`. The trouble here is we don't have a good
  incremental way for customers to try out the new language support before upgrading, given the
  major changes between 2.x and 3.0. So the backup plan if a customer experiences major code
  intelligence regressions on upgrade is this: frantically convert `java-langserver` to work with
  websockets and raw API. Estimate can get this up in 2-3 full days of coding.
- Edge cases in build systems. Maven and Gradle, but Gradle especially. Also unclear if the changes
  we make will be accepted upstream. Will try to avoid too much "magic", but that may mean some
  customers will need to add manual config to repositories that were formerly supported
  automatically.
- Performance issues with upfront repo fetching. The old `java-langserver` fetched files on the
  fly. At least some of this work targeted performance issues experienced by customers. It seems now
  after investigation and [feedback from the eclipse.jdt.ls
  team](https://github.com/eclipse/eclipse.jdt.ls/issues/905) that supporting this in
  `eclipse.jdt.ls` will take awhile (both due to technical difficulty and getting input and approval
  from Eclipse maintainers). Will mitigate this by testing on large repositories in the first
  milestone. The outcome of that testing will dictate how much extra work we'll need to do on this
  front.
