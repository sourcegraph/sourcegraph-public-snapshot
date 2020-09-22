
<!-- BEGIN ASSIGNEE: aidaeology -->
@aidaeology

- [ ] ~update lsif.dev~ ([#12501](https://github.com/sourcegraph/sourcegraph/issues/12501)) 
- [ ] ~Make bundle manager janitor more lenient during disaster recovery~ ([#12168](https://github.com/sourcegraph/sourcegraph/issues/12168)) 🧶
- [ ] ~Remove -endpoint from src-cli usage~ ([#11878](https://github.com/sourcegraph/sourcegraph/issues/11878)) 🧶
- [ ] ~Programatically create regular data about code intel activity~ ([#11746](https://github.com/sourcegraph/sourcegraph/issues/11746)) 

Completed
- [x] Create quick-start video for setting up Code Intelligence in CI/CD ([~#13049~](https://github.com/sourcegraph/sourcegraph/issues/13049)) 
- [x] Research language usage among customers and projects ([~#12726~](https://github.com/sourcegraph/sourcegraph/issues/12726)) 
- [x] Determine state of LSIF-java ([~#12306~](https://github.com/sourcegraph/sourcegraph/issues/12306)) 
- [x] Code Intelligence 3.19 Tracking issue ([~#12132~](https://github.com/sourcegraph/sourcegraph/issues/12132)) 
- [x] Update left side navigation [#13188](https://github.com/sourcegraph/sourcegraph/pull/13188) :shipit:
- [x] Edit of language guide page to improve phrasing [#13108](https://github.com/sourcegraph/sourcegraph/pull/13108) :shipit:
- [x] Updated adding lsif to workflows docs [#13078](https://github.com/sourcegraph/sourcegraph/pull/13078) :shipit:
- [x] Improving code intelligence docs and LSIF indexer installation instructions [#13041](https://github.com/sourcegraph/sourcegraph/pull/13041) :shipit:
- [x] Added goals to code intel team page [#1369](https://github.com/sourcegraph/about/pull/1369) :shipit:
- [x] Updated code intel page with NEO [#1327](https://github.com/sourcegraph/about/pull/1327) :shipit:
<!-- END ASSIGNEE -->

<!-- BEGIN ASSIGNEE: efritz -->
@efritz

- [ ] ~codeintel: Alert on failure to open database~ ([#12712](https://github.com/sourcegraph/sourcegraph/issues/12712)) 🧶

Completed
- [x] Create quick-start video for setting up Code Intelligence in CI/CD ([~#13049~](https://github.com/sourcegraph/sourcegraph/issues/13049)) 
- [x] RFC 199: Build indexer service ([~#12707~](https://github.com/sourcegraph/sourcegraph/issues/12707); PRs: ~[#12723](https://github.com/sourcegraph/sourcegraph/pull/12723)~) 
- [x] RFC 199: Expose internal routes for the indexer ([~#12666~](https://github.com/sourcegraph/sourcegraph/issues/12666); PRs: ~[#12691](https://github.com/sourcegraph/sourcegraph/pull/12691)~) 
- [x] RFC 199: Index queue client ([~#12665~](https://github.com/sourcegraph/sourcegraph/issues/12665); PRs: ~[#12688](https://github.com/sourcegraph/sourcegraph/pull/12688)~) 
- [x] RFC 199: Index queue API ([~#12664~](https://github.com/sourcegraph/sourcegraph/issues/12664); PRs: ~[#12657](https://github.com/sourcegraph/sourcegraph/pull/12657)~) 
- [x] Query definition and hover provider so we can correctly badge results ([~#12133~](https://github.com/sourcegraph/sourcegraph/issues/12133)) 
- [x] Code Intelligence 3.19 Tracking issue ([~#12132~](https://github.com/sourcegraph/sourcegraph/issues/12132)) 
- [x] Implement better nearest commit queries ([~#12098~](https://github.com/sourcegraph/sourcegraph/issues/12098); PRs: ~[#12422](https://github.com/sourcegraph/sourcegraph/pull/12422)~, ~[#12411](https://github.com/sourcegraph/sourcegraph/pull/12411)~, ~[#12408](https://github.com/sourcegraph/sourcegraph/pull/12408)~, ~[#12406](https://github.com/sourcegraph/sourcegraph/pull/12406)~, ~[#12404](https://github.com/sourcegraph/sourcegraph/pull/12404)~, ~[#12402](https://github.com/sourcegraph/sourcegraph/pull/12402)~, ~[#12401](https://github.com/sourcegraph/sourcegraph/pull/12401)~) 
- [x] Visible at tip calculation is racy ([~#12095~](https://github.com/sourcegraph/sourcegraph/issues/12095); PRs: ~[#12422](https://github.com/sourcegraph/sourcegraph/pull/12422)~) 🐛
- [x] ~RFC 199: Deploy new indexer service~ ([~#12709~](https://github.com/sourcegraph/sourcegraph/issues/12709)) 
- [x] ~RFC 199: Replace docker commands with firecracker commands~ ([~#12708~](https://github.com/sourcegraph/sourcegraph/issues/12708)) __5d__ 
- [x] ~UI Tooltips for asymmetric precision of hover and definition results~ ([~#12706~](https://github.com/sourcegraph/sourcegraph/issues/12706)) __0.5d__ 
- [x] codeintel: Install and use the LSIF upload route on the codeintel internal API [#13157](https://github.com/sourcegraph/sourcegraph/pull/13157) :shipit:
- [x] Add -upload-route flag to lsif upload [#267](https://github.com/sourcegraph/src-cli/pull/267) :shipit:
- [x] codeintel: Install docker inside precise-code-intel-indexer-vm docker image [#13119](https://github.com/sourcegraph/sourcegraph/pull/13119) :shipit:
- [x] docker-images: Add ignite-ubuntu [#12919](https://github.com/sourcegraph/sourcegraph/pull/12919) :shipit:
- [x] codeintel: Add fast-path edge unmarshalling [#12878](https://github.com/sourcegraph/sourcegraph/pull/12878) :shipit:🧶
- [x] codeintel: Add disable indexer flag [#12800](https://github.com/sourcegraph/sourcegraph/pull/12800) :shipit:
- [x] codeintel: Add mocks for queue client [#12798](https://github.com/sourcegraph/sourcegraph/pull/12798) :shipit:
- [x] internal API proxy: Remove verb allowlist in gitservice proxy [#12797](https://github.com/sourcegraph/sourcegraph/pull/12797) :shipit:
- [x] codeintel: Collapse worker handler and processor [#12795](https://github.com/sourcegraph/sourcegraph/pull/12795) :shipit:
- [x] workerutil: Make generic store [#12792](https://github.com/sourcegraph/sourcegraph/pull/12792) :shipit:
- [x] Fix default host for SRC_HTTP_ADDR_INTERNAL [#12768](https://github.com/sourcegraph/sourcegraph/pull/12768) :shipit:
- [x] Revert "Revert "codeintel: Internal API proxy (#12691)" (#12728)" [#12758](https://github.com/sourcegraph/sourcegraph/pull/12758) :shipit:
- [x] codeintel: VM-based indexer service [#12723](https://github.com/sourcegraph/sourcegraph/pull/12723) :shipit:
- [x] codeintel: Internal API proxy [#12691](https://github.com/sourcegraph/sourcegraph/pull/12691) :shipit:
- [x] codeintel: Index queue client [#12688](https://github.com/sourcegraph/sourcegraph/pull/12688) :shipit:
- [x] workerutil: Move store into own package [#12663](https://github.com/sourcegraph/sourcegraph/pull/12663) :shipit:
- [x] workerutil: Add DequeueWithIndependentTransactionContext to store [#12661](https://github.com/sourcegraph/sourcegraph/pull/12661) :shipit:
- [x] codeintel: Index queue API [#12657](https://github.com/sourcegraph/sourcegraph/pull/12657) :shipit:
- [x] Remove committed binary [#12594](https://github.com/sourcegraph/sourcegraph/pull/12594) :shipit:
- [x] Add index on repo name column [#12591](https://github.com/sourcegraph/sourcegraph/pull/12591) :shipit:
- [x] Categorize enterprise frontend startup behaviors [#12539](https://github.com/sourcegraph/sourcegraph/pull/12539) :shipit:
- [x] Move [enterprise/]cmd/frontend/authz to [enterprise/]internal/authz [#12538](https://github.com/sourcegraph/sourcegraph/pull/12538) :shipit:
- [x] codeintel: Remove ErrMalformedBundle [#12497](https://github.com/sourcegraph/sourcegraph/pull/12497) :shipit:
- [x] codeintel: Remove unused store, gitserver code [#12425](https://github.com/sourcegraph/sourcegraph/pull/12425) :shipit:
- [x]  codeintel: Update store to target new nearest upload tables [#12422](https://github.com/sourcegraph/sourcegraph/pull/12422) :shipit:
- [x] codeintel: Add commit updater utility [#12411](https://github.com/sourcegraph/sourcegraph/pull/12411) :shipit:
- [x] codeintel: Add CalculateVisibleUploads to store [#12408](https://github.com/sourcegraph/sourcegraph/pull/12408) :shipit:
- [x] workerutil: Fix nil-deref in when using worker store transactionally [#12407](https://github.com/sourcegraph/sourcegraph/pull/12407) :shipit:
- [x] codeintel: Add tables for denormalizing nearest upload data [#12406](https://github.com/sourcegraph/sourcegraph/pull/12406) :shipit:
- [x] codeintel: Add CommitGraph to gitserver client [#12404](https://github.com/sourcegraph/sourcegraph/pull/12404) :shipit:
- [x] codeintel: Add commit graph utilities [#12402](https://github.com/sourcegraph/sourcegraph/pull/12402) :shipit:
- [x] codeintel: Add Lock to store [#12401](https://github.com/sourcegraph/sourcegraph/pull/12401) :shipit:
- [x] workerutil: Make all column names customizable [#12398](https://github.com/sourcegraph/sourcegraph/pull/12398) :shipit:
- [x] codeintel: Apply rate limit to gitserver requests from indexability scheduler [#12379](https://github.com/sourcegraph/sourcegraph/pull/12379) :shipit:
- [x] codeintel: Group code intel data for serialization on-demand [#12125](https://github.com/sourcegraph/sourcegraph/pull/12125) :shipit:
- [x] codeintel: Additional worker memory improvements [#12108](https://github.com/sourcegraph/sourcegraph/pull/12108) :shipit:
<!-- END ASSIGNEE -->

<!-- BEGIN ASSIGNEE: gbrik -->
@gbrik

- [ ] ~Investigate effort for a new LSIF-swift indexer~ ([#12350](https://github.com/sourcegraph/sourcegraph/issues/12350)) 🕵️
- [ ] ~🚚 LSIF-clang Delivery~ ([#12349](https://github.com/sourcegraph/sourcegraph/issues/12349)) __4d__ 
- [ ] ~Create a code intel user survey~ ([#11747](https://github.com/sourcegraph/sourcegraph/issues/11747)) 
- [ ] ~Programatically create regular data about code intel activity~ ([#11746](https://github.com/sourcegraph/sourcegraph/issues/11746)) 

Completed
- [x] Code Intelligence 3.19 Tracking issue ([~#12132~](https://github.com/sourcegraph/sourcegraph/issues/12132)) 
<!-- END ASSIGNEE -->
