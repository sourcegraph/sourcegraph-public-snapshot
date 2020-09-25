
<!-- BEGIN ASSIGNEE: Strum355 -->
@Strum355

- [ ] 🚚 LSIF-Java Delivery  ([#13017](https://github.com/sourcegraph/sourcegraph/issues/13017)) 
- [ ] Find references to common Java class method name without false-positives ([#3418](https://github.com/sourcegraph/sourcegraph/issues/3418)) 

Completed
- [x] (🏁 today) LSIF-JVM Research Spike ([~#14057~](https://github.com/sourcegraph/sourcegraph/issues/14057)) 🕵️
<!-- END ASSIGNEE -->

<!-- BEGIN ASSIGNEE: aidaeology -->
@aidaeology: __5.00d__

- [ ] Create successful and reproducible indexes of 20 OSS repos ([#12](https://github.com/sourcegraph/lsif-clang/issues/12)) 
- [ ] 🚚 LSIF-Go Delivery ([#13015](https://github.com/sourcegraph/sourcegraph/issues/13015)) __5d__ 

Completed
- [x] (🏁 today) Add history of code intelligence (~[#1628](https://github.com/sourcegraph/about/pull/1628)~) :shipit:
<!-- END ASSIGNEE -->

<!-- BEGIN ASSIGNEE: efritz -->
@efritz: __11.00d__

- [ ] Write architecture docs for precise code intel indexer ([#14010](https://github.com/sourcegraph/sourcegraph/issues/14010)) 
- [ ] Write announcement post for Postgres change ([#14009](https://github.com/sourcegraph/sourcegraph/issues/14009)) 
- [ ] Update RFC 236 ([#14007](https://github.com/sourcegraph/sourcegraph/issues/14007)) 
- [ ] RFC 235: Tracking issue ([#13882](https://github.com/sourcegraph/sourcegraph/issues/13882)) __5.50d__ 
  - [x] (🏁 6 days ago) RFC 235: Add code intel postgres image ([~#13912~](https://github.com/sourcegraph/sourcegraph/issues/13912); PRs: ~[#13913](https://github.com/sourcegraph/sourcegraph/pull/13913)~) __0.5d__ 
  - [ ] RFC 235: Cleanup ([#13890](https://github.com/sourcegraph/sourcegraph/issues/13890)) __1d__ 
  - [ ] RFC 235: Update worker to write to Postgres ([#13889](https://github.com/sourcegraph/sourcegraph/issues/13889); PRs: [#13946](https://github.com/sourcegraph/sourcegraph/pull/13946), [#13923](https://github.com/sourcegraph/sourcegraph/pull/13923)) __0.5d__ 
  - [ ] RFC 235: Migrate SQLite data to Postgres ([#13888](https://github.com/sourcegraph/sourcegraph/issues/13888); PRs: [#13932](https://github.com/sourcegraph/sourcegraph/pull/13932), [#13923](https://github.com/sourcegraph/sourcegraph/pull/13923)) __0.5d__ 
  - [ ] RFC 235: Update bundle manager to read from Postgres ([#13886](https://github.com/sourcegraph/sourcegraph/issues/13886); PRs: [#13924](https://github.com/sourcegraph/sourcegraph/pull/13924), [#13923](https://github.com/sourcegraph/sourcegraph/pull/13923)) __0.5d__ 
  - [ ] RFC 235: Add migration infrastructure to codeintel database ([#13885](https://github.com/sourcegraph/sourcegraph/issues/13885); PRs: ~[#13943](https://github.com/sourcegraph/sourcegraph/pull/13943)~, [#13903](https://github.com/sourcegraph/sourcegraph/pull/13903)) __1d__ 
  - [ ] RFC 235: Configure connection to codeintel database ([#13884](https://github.com/sourcegraph/sourcegraph/issues/13884); PRs: ~[#13952](https://github.com/sourcegraph/sourcegraph/pull/13952)~, [#13864](https://github.com/sourcegraph/sourcegraph/pull/13864)) __0.5d__ 
  - [ ] RFC 235: Add code intel postgres container ([#13883](https://github.com/sourcegraph/sourcegraph/issues/13883); PRs: [#13924](https://github.com/sourcegraph/sourcegraph/pull/13924), [#13904](https://github.com/sourcegraph/sourcegraph/pull/13904), [#13864](https://github.com/sourcegraph/sourcegraph/pull/13864)) __1d__ 
  - [x] (🏁 6 days ago) chore: Set exec bit on docker-images/codeintel-db/build.sh (~[#13955](https://github.com/sourcegraph/sourcegraph/pull/13955)~) :shipit:
  - [x] (🏁 6 days ago) chore: Co-locate dev scripts to interact with postgres (~[#13942](https://github.com/sourcegraph/sourcegraph/pull/13942)~) :shipit:
  - [ ] codeintel: RFC 235: Soft delete upload records ([#13822](https://github.com/sourcegraph/sourcegraph/pull/13822)) :shipit:
- [ ] RFC 201: Tracking issue ([#13891](https://github.com/sourcegraph/sourcegraph/issues/13891)) 
  - [ ] RFC 201: Use auto-configurator ([#13898](https://github.com/sourcegraph/sourcegraph/issues/13898)) 
  - [ ] RFC 201: Write auto-configurator ([#13897](https://github.com/sourcegraph/sourcegraph/issues/13897)) 
  - [ ] RFC 201: Create index record based on configuration file ([#13896](https://github.com/sourcegraph/sourcegraph/issues/13896)) 
  - [ ] RFC 201: Update index scheduler to publish new payload ([#13895](https://github.com/sourcegraph/sourcegraph/issues/13895)) 
  - [ ] RFC 201: Update auto indexer execution ([#13894](https://github.com/sourcegraph/sourcegraph/issues/13894)) 
  - [ ] RFC 201: Write a JSONC index configuration format parser ([#13892](https://github.com/sourcegraph/sourcegraph/issues/13892)) 
  - [ ] codeintel: Make docker/firecracker abstraction in indexer ([#14121](https://github.com/sourcegraph/sourcegraph/pull/14121)) :shipit:
  - [x] (🏁 today) codeintel: Add orderedKeys to indexer (~[#14117](https://github.com/sourcegraph/sourcegraph/pull/14117)~) :shipit:
  - [x] (🏁 today) codeintel: Refactor construction of copyfiles flags (~[#14116](https://github.com/sourcegraph/sourcegraph/pull/14116)~) :shipit:
  - [x] (🏁 today) codeintel: Move away from hardcoding image names (~[#14114](https://github.com/sourcegraph/sourcegraph/pull/14114)~) :shipit:
  - [x] (🏁 today) codeintel: Refactor index command construction (~[#14105](https://github.com/sourcegraph/sourcegraph/pull/14105)~) :shipit:
  - [x] (🏁 today) codeintel: Lower indexer output to debug level (~[#14103](https://github.com/sourcegraph/sourcegraph/pull/14103)~) :shipit:
  - [x] (🏁 1 day ago) codeintel: Refactor command runner in indexer (~[#14102](https://github.com/sourcegraph/sourcegraph/pull/14102)~) :shipit:
- [ ] 🚚 LSIF-Go Delivery ([#13015](https://github.com/sourcegraph/sourcegraph/issues/13015)) __5d__ 
- [ ] codeintel: No longer able to upload repos which are currently cloning ([#14052](https://github.com/sourcegraph/sourcegraph/issues/14052); PRs: [#14141](https://github.com/sourcegraph/sourcegraph/pull/14141)) 🐛
- [ ] codeintel: git diffing fails graphql requests related to force-pushed commits ([#12588](https://github.com/sourcegraph/sourcegraph/issues/12588)) 🧶
- [ ] chore: Fix tracking-issue tests ([#14172](https://github.com/sourcegraph/sourcegraph/pull/14172)) :shipit:
- [ ] codeintel: Remove some wrappers from a previous abstraction ([#14142](https://github.com/sourcegraph/sourcegraph/pull/14142)) :shipit:

Completed: __1.00d__
- [x] (🏁 6 days ago) chore: Co-locate dev scripts to interact with postgres (~[#13942](https://github.com/sourcegraph/sourcegraph/pull/13942)~) :shipit:
- [x] (🏁 6 days ago) chore: Multiple database handles (~[#13952](https://github.com/sourcegraph/sourcegraph/pull/13952)~) :shipit:
- [x] (🏁 6 days ago) RFC 235: Add code intel postgres image ([~#13912~](https://github.com/sourcegraph/sourcegraph/issues/13912); PRs: ~[#13913](https://github.com/sourcegraph/sourcegraph/pull/13913)~) __0.5d__ 
- [x] (🏁 6 days ago) chore: Set exec bit on docker-images/codeintel-db/build.sh (~[#13955](https://github.com/sourcegraph/sourcegraph/pull/13955)~) :shipit:
- [x] (🏁 6 days ago) chore: Relocate frontend migrations (~[#13943](https://github.com/sourcegraph/sourcegraph/pull/13943)~) :shipit:
- [x] (🏁 2 days ago) tracking-issue: Fix timeout (~[#13986](https://github.com/sourcegraph/sourcegraph/pull/13986)~) :shipit:
- [x] (🏁 2 days ago) tracking-issue: Fix long-form linked issues (~[#13989](https://github.com/sourcegraph/sourcegraph/pull/13989)~) :shipit:
- [x] (🏁 2 days ago) 504 Gateway Timeouts when mousing over after the page has loaded for a while ([~#12930~](https://github.com/sourcegraph/sourcegraph/issues/12930)) 🐛
- [x] (🏁 2 days ago) LSIF uploads fail with abbreviated OID ([~#13957~](https://github.com/sourcegraph/sourcegraph/issues/13957); PRs: ~[#14005](https://github.com/sourcegraph/sourcegraph/pull/14005)~) __0.5d__ 
- [x] (🏁 2 days ago) tracking-issue: Nested tracking issues (~[#13998](https://github.com/sourcegraph/sourcegraph/pull/13998)~) :shipit:
- [x] (🏁 2 days ago) Fix retries in src-cli lsif upload ([~#14008~](https://github.com/sourcegraph/sourcegraph/issues/14008)) 
- [x] (🏁 2 days ago) tracking-issue: Parallelize writes (~[#14006](https://github.com/sourcegraph/sourcegraph/pull/14006)~) :shipit:
- [x] (🏁 2 days ago) tracking-issue: Break code up into files according to type/action (~[#14013](https://github.com/sourcegraph/sourcegraph/pull/14013)~) :shipit:
- [x] (🏁 2 days ago) tracking-issue: List PRs for an issue inline (~[#14018](https://github.com/sourcegraph/sourcegraph/pull/14018)~) :shipit:
- [x] (🏁 2 days ago) tracking-issue: Separate complete/incomplete work (~[#14034](https://github.com/sourcegraph/sourcegraph/pull/14034)~) :shipit:
- [x] (🏁 2 days ago) tracking-issue: Better nested tracking issue estimates (~[#14035](https://github.com/sourcegraph/sourcegraph/pull/14035)~) :shipit:
- [x] (🏁 1 day ago) codenotify: Configure efritz's subscriptions (~[#14060](https://github.com/sourcegraph/sourcegraph/pull/14060)~) :shipit:
- [x] (🏁 1 day ago) codeintel: Refactor command runner in indexer (~[#14102](https://github.com/sourcegraph/sourcegraph/pull/14102)~) :shipit:
- [x] (🏁 today) codeintel: Lower indexer output to debug level (~[#14103](https://github.com/sourcegraph/sourcegraph/pull/14103)~) :shipit:
- [x] (🏁 today) tracking-issue: Fix strikethrough on closed (un-merged) PRs (~[#14107](https://github.com/sourcegraph/sourcegraph/pull/14107)~) :shipit:
- [x] (🏁 today) codeintel: Refactor index command construction (~[#14105](https://github.com/sourcegraph/sourcegraph/pull/14105)~) :shipit:
- [x] (🏁 today) codeintel: Move away from hardcoding image names (~[#14114](https://github.com/sourcegraph/sourcegraph/pull/14114)~) :shipit:
- [x] (🏁 today) codeintel: Refactor construction of copyfiles flags (~[#14116](https://github.com/sourcegraph/sourcegraph/pull/14116)~) :shipit:
- [x] (🏁 today) codeintel: Add orderedKeys to indexer (~[#14117](https://github.com/sourcegraph/sourcegraph/pull/14117)~) :shipit:
- [x] (🏁 today) tracking-issue: Nest unlinked PRs under the closest tracking issue (~[#14108](https://github.com/sourcegraph/sourcegraph/pull/14108)~) :shipit:
- [x] (🏁 today) dbworker: Pass sql options to TransactableHandle ([~#14044~](https://github.com/sourcegraph/sourcegraph/issues/14044); PRs: ~[#14063](https://github.com/sourcegraph/sourcegraph/pull/14063)~, ~[#14061](https://github.com/sourcegraph/sourcegraph/pull/14061)~) 
- [x] (🏁 today) tracking-issue: Order finished work chronologically (~[#14124](https://github.com/sourcegraph/sourcegraph/pull/14124)~) :shipit:
- [x] (🏁 today) tracking-issue: Fix tests (~[#14168](https://github.com/sourcegraph/sourcegraph/pull/14168)~) :shipit:
- [x] (🏁 today) tracking-issue: Do not show completed PRs if the owning image is also complete (~[#14169](https://github.com/sourcegraph/sourcegraph/pull/14169)~) :shipit:
<!-- END ASSIGNEE -->

<!-- BEGIN ASSIGNEE: gbrik -->
@gbrik: __1.00d__

- [ ] Investigate effort for Bazel integration ([#13202](https://github.com/sourcegraph/sourcegraph/issues/13202)) __1d__ 🕵️
- [ ] no output produced for seemingly well-formed compile_commands.json ([#4](https://github.com/sourcegraph/lsif-clang/issues/4)) 🐛
- [ ] doesn't work on arch linux ([#1](https://github.com/sourcegraph/lsif-clang/issues/1)) 🐛
- [ ] github.com/gabime/spdlog doesn't produce LSIF output for template definitions ([#14](https://github.com/sourcegraph/lsif-clang/issues/14)) 🐛
- [ ] github.com/nlohmann/json on macOS fails ([#13](https://github.com/sourcegraph/lsif-clang/issues/13)) 🐛
- [ ] infer project root automatically if not specified ([#15](https://github.com/sourcegraph/lsif-clang/issues/15)) 
- [ ] Create successful and reproducible indexes of 20 OSS repos ([#12](https://github.com/sourcegraph/lsif-clang/issues/12)) 
<!-- END ASSIGNEE -->
