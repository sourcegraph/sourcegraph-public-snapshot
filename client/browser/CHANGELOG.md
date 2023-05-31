<!--
###################################### READ ME ###########################################
### This changelog should always be read on `main` branch. Its contents on version   ###
### branches do not necessarily reflect the changes that have gone into that branch.   ###
##########################################################################################
-->

# Changelog

All notable changes to Sourcegraph [Browser Extensions](./README.md) are documented in this file.

<!-- START CHANGELOG -->

## Unreleased

- add ability to remove saved sg url from suggestion list [pull/52555](https://github.com/sourcegraph/sourcegraph/pull/52555)

- Fix code-intel tooltips for pull request pages on Bitbucket: https://github.com/sourcegraph/sourcegraph/pull/52609

## Chrome & Firefox 23.4.14.1343, Safari 1.24

- Fix view on Sourcegraph links on GitHub (global navigation update feature enabled): https://github.com/sourcegraph/sourcegraph/pull/50551
- Bump code-intel extensions bundles version: https://github.com/sourcegraph/sourcegraph/pull/50631

## Chrome & Firefox 23.3.10.1712, Safari 1.23

- Fix code intelligence popup actions: [pull/49025](https://github.com/sourcegraph/sourcegraph/pull/49025)

## Chrome & Firefox 23.3.8.2218, Safari 1.22

- Fix code intelligence popup actions: [issues/48918](https://github.com/sourcegraph/sourcegraph/issues/48918)

## Chrome 23.3.1.1624, Firefox 23.3.1.1623, Safari 1.21

- Remove deprecated code-intel endpoints usage.

## Chrome 23.2.17.1613, Firefox 23.2.17.1612, Safari 1.20

- Omit credentials in legacy extension bundles requests: [pull/47486](https://github.com/sourcegraph/sourcegraph/pull/47486)
- Fix selectors used to inject Sourcegraph code intelligence on GitHub: [pull/47427](https://github.com/sourcegraph/sourcegraph/pull/47427)

## Chrome & Firefox 22.11.24.1820, Safari v1.19

- Fix route change handlers for GitHub: [issues/44074](https://github.com/sourcegraph/sourcegraph/issues/44074), [pull/44783](https://github.com/sourcegraph/sourcegraph/pull/44783)
- Fix selectors used to inject Sourcegraph code intelligence on GitHub: [issues/44759](https://github.com/sourcegraph/sourcegraph/issues/44759), [pull/44766](https://github.com/sourcegraph/sourcegraph/pull/44766)

## Chrome 22.11.14.1520, Firefox 22.11.14.1521, Safari v1.18

- Add new GitHub UI support: [issues/44237](https://github.com/sourcegraph/sourcegraph/issues/44237), [pull/44285](https://github.com/sourcegraph/sourcegraph/pull/44285)
- Remove 'Search on Sourcegraph' buttons from GitHub search pages: [pull/44328](https://github.com/sourcegraph/sourcegraph/pull/44328)

## Chrome 22.10.18.1144, Firefox 22.10.18.1133, Safari v1.17

- Updated the info text when accessing an unindexed repository [pull/42509](https://github.com/sourcegraph/sourcegraph/pull/42509)
- Fix an issue that caused the native code host integration to not work on gitlab.com [pull/42748](https://github.com/sourcegraph/sourcegraph/pull/42748)

## Chrome & Firefox v22.9.27.1330, Safari v1.16

- Remove tracking parameters from all shareable URLs [pull/42022](https://github.com/sourcegraph/sourcegraph/pull/42022)

## Chrome v22.9.14.1335, Firefox v22.9.14.1616, Safari v1.15

- Fix code-intel issue on GitHub Enterprise: [pull/41646](https://github.com/sourcegraph/sourcegraph/pull/41646)

## Chrome & Firefox 22.7.29.851, Safari v1.14 (22.8.2.1319)

- Fix extensions decorations issue when navigating project files on GitHub: [codecov/sourcegraph-codecov/issues/86](https://github.com/codecov/sourcegraph-codecov/issues/86), [pull/39557](https://github.com/sourcegraph/sourcegraph/pull/39557)
- Fix command palette style regression on GitHub: [issues/39495](https://github.com/sourcegraph/sourcegraph/issues/39495), [issues/33433](https://github.com/sourcegraph/sourcegraph/issues/33433) [pull/39580](https://github.com/sourcegraph/sourcegraph/pull/39580)

## Chrome & Firefox v22.7.11.926

- Fix Sourcegraph buttons styles on Bitbucket cloud: [issues/32598](https://github.com/sourcegraph/sourcegraph/issues/32598), [pull/33787](https://github.com/sourcegraph/sourcegraph/pull/33787)
- Fix repo visibility check logic: [issues/29244](https://github.com/sourcegraph/sourcegraph/issues/29244), [pull/33352](https://github.com/sourcegraph/sourcegraph/pull/33352)
- Add different browser extension icons for development mode builds: [issue/33587](https://github.com/sourcegraph/sourcegraph/issues/33587), [pull/34353](https://github.com/sourcegraph/sourcegraph/pull/34353)
- Fix git-extras extension blame for selected lines issue on the code hosts: [issues/34700](https://github.com/sourcegraph/sourcegraph/issues/34700), [pull/34698](https://github.com/sourcegraph/sourcegraph/pull/34698)
- Fix hover-overlay styling issue on GitLab: [issues/35315](https://github.com/sourcegraph/sourcegraph/issues/35315), [pull/35403](https://github.com/sourcegraph/sourcegraph/pull/35403)

## Chrome & Firefox v22.4.7.1712, Safari v1.13

- Update banners for not synced private repositories banners when on Sourcegraph Cloud instance and not added repositories when on other instances: [pull/31922](https://github.com/sourcegraph/sourcegraph/pull/31922), [issues/31920](https://github.com/sourcegraph/sourcegraph/issues/31920)
- Fix style errors in browser console: [pull/32604](https://github.com/sourcegraph/sourcegraph/pull/32604), [issues/32443](https://github.com/sourcegraph/sourcegraph/issues/32443)
- Fix styles conflict on GitLab: [pull/32548](https://github.com/sourcegraph/sourcegraph/pull/32548), [issues/32462](https://github.com/sourcegraph/sourcegraph/issues/32462)
- Fixes telemetry for initial "browser extension installed" event [#pull/33175](https://github.com/sourcegraph/sourcegraph/pull/33175), [#issues/33143](https://github.com/sourcegraph/sourcegraph/issues/33143)
- Fix native integration to pass cookies/credentials for "Extensions" info GraphQL request [#pull/33406](https://github.com/sourcegraph/sourcegraph/pull/33406), [#issues/32599](https://github.com/sourcegraph/sourcegraph/issues/32599)

## Chrome & Firefox v22.3.11.1145, Safari v1.12

- Fix client-side routing support on GitHub repository browse file tree pages: [#pull/32199](https://github.com/sourcegraph/sourcegraph/pull/32199), [#issues/31716](https://github.com/sourcegraph/sourcegraph/issues/31716)
- Fix code intel popup buttons overflow issue on GitHub: [#pull/31698](https://github.com/sourcegraph/sourcegraph/pull/31698), [#issues/31359](https://github.com/sourcegraph/sourcegraph/issues/31359)
- Make 'Configure on Sourcegraph' button navigate to manage repositories page when on default Sourcegraph URL: [#pull/31690](https://github.com/sourcegraph/sourcegraph/pull/31690), [#issues/3066](https://github.com/sourcegraph/sourcegraph/issues/3066)
- Add installs/uninstall events tracking: [#pull/31785](https://github.com/sourcegraph/sourcegraph/pull/31785), [issues/31486](https://github.com/sourcegraph/sourcegraph/issues/31486)

## Chrome & Firefox v22.2.11.1553, Safari v1.11

- Make "single click to definition" an opt-in through advanced settings: [#pull/30540](https://github.com/sourcegraph/sourcegraph/pull/30540), [#issues/#30437](https://github.com/sourcegraph/sourcegraph/issues/30437)
- Add "https://" URL input placeholder [#pull/30282](https://github.com/sourcegraph/sourcegraph/pull/30282), [#issues/14723](https://github.com/sourcegraph/sourcegraph/issues/14723)
- Add filtering of browser extension dropdown duplicated URLs [#pull/30674](https://github.com/sourcegraph/sourcegraph/pull/30674), [#issues/30673](https://github.com/sourcegraph/sourcegraph/issues/30673)
- Add code intel support to GitHub pull request commit view [#pull/30618](https://github.com/sourcegraph/sourcegraph/pull/30618), [#issues/30623](https://github.com/sourcegraph/sourcegraph/issues/30623)
- Add tracking of inbound traffic from browser extension/code host integration [#pull/30170](https://github.com/sourcegraph/sourcegraph/pull/30170), [#issues/27082](https://github.com/sourcegraph/sourcegraph/issues/27082)
- Add "Search on Sourcegraph" buttons to GitHub search pages [pull/#30399](https://github.com/sourcegraph/sourcegraph/pull/30399), [#issues/10410](https://github.com/sourcegraph/sourcegraph/issues/10410), [#issues/30968](https://github.com/sourcegraph/sourcegraph/issues/30968)

## Chrome & Firefox v22.1.25.1535, Safari v1.10

- Add extra field to log browser extension version [#issues/27845](https://github.com/sourcegraph/sourcegraph/issues/27845), [pull/27902](https://github.com/sourcegraph/sourcegraph/pull/27902)
- Implement Sourcegraph URL dropdown for ease of URL switching [#issues/29030](https://github.com/sourcegraph/sourcegraph/issues/29030), [#pull/29471](https://github.com/sourcegraph/sourcegraph/pull/29471)
  - Fix incorrect event log URL detection. Allow event logs for non-private repositories. [#issues/25778](https://github.com/sourcegraph/sourcegraph/issues/25778)

## Chrome v21.12.10.1012, Firefox v21.12.10.1048, Safari v1.9

- Private cloud repositories should use hashed identifiers instead of repository names [pull/28621](https://github.com/sourcegraph/sourcegraph/pull/28621), [#pull/28387](https://github.com/sourcegraph/sourcegraph/pull/28387), [#issues/27922](https://github.com/sourcegraph/sourcegraph/issues/27922)

  - Also, browser extension does not override native codehost tooltips whenever cannot inject code intelligence
  - Also, browser extension shows "Sign In to Sourcegraph" button on private repository pages

- Browser Extension Telemetry documentation [#pull/28689](https://github.com/sourcegraph/sourcegraph/pull/28689), [#issues/27383](https://github.com/sourcegraph/sourcegraph/issues/27383)

## Chrome / Firefox v21.11.25.1400, Safari v1.8

- Fix omnibox opening in wrong tab [#issues/23475](https://github.com/sourcegraph/sourcegraph/issues/23475), [#pull/27525](https://github.com/sourcegraph/sourcegraph/pull/27525)
- Fix excessive "all websites" permissions for Safari [#issues/23542](https://github.com/sourcegraph/sourcegraph/issues/23542), [#pull/26832](https://github.com/sourcegraph/sourcegraph/pull/26832)
- Update private repository detection logic for Gitlab and Github [#issues/27382](https://github.com/sourcegraph/sourcegraph/issues/27382), [#pull/27779](https://github.com/sourcegraph/sourcegraph/pull/27779)
- Fix open file diff bug for Sourcegraph URL with trailing slash [#26832](https://github.com/sourcegraph/sourcegraph/pull/28058)
- Disable (temporarily) browser extension for private repositories when using Sourcegraph Cloud URL [#issues/28070](https://github.com/sourcegraph/sourcegraph/issues/28070), [#pull/28089](https://github.com/sourcegraph/sourcegraph/pull/28089)
- Fix code host tooltip displaying issue [#issues/19560](https://github.com/sourcegraph/sourcegraph/issues/19560), [#pull/27381](https://github.com/sourcegraph/sourcegraph/pull/27381)

- Fix omnibox opening in wrong tab [#issues/23475](https://github.com/sourcegraph/sourcegraph/issues/23475), [#pull/27525](https://github.com/sourcegraph/sourcegraph/pull/27525)
- Fix excessive "all websites" permissions for Safari [#issues/23542](https://github.com/sourcegraph/sourcegraph/issues/23542), [#pull/26832](https://github.com/sourcegraph/sourcegraph/pull/26832)
- Update private repository detection logic for Gitlab and Github [#issues/27382](https://github.com/sourcegraph/sourcegraph/issues/27382), [#pull/27779](https://github.com/sourcegraph/sourcegraph/pull/27779)
- Fix open file diff bug for Sourcegraph URL with trailing slash [#26832](https://github.com/sourcegraph/sourcegraph/pull/28058)
- Disable (temporarily) browser extension for private repositories when using Sourcegraph Cloud URL [#issues/28070](https://github.com/sourcegraph/sourcegraph/issues/28070), [#pull/28089](https://github.com/sourcegraph/sourcegraph/pull/28089)

## Chrome / Firefox v21.11.8.1804, Safari v1.7

- URL mismatch [#issues/26170](https://github.com/sourcegraph/sourcegraph/issues/26170), [#pull/26776](https://github.com/sourcegraph/sourcegraph/pull/26776)
- Branded hover [#issues/23738](https://github.com/sourcegraph/sourcegraph/issues/23738), [#pull/26189](https://github.com/sourcegraph/sourcegraph/pull/26189)
- Action icon quality [#issues/16218](https://github.com/sourcegraph/sourcegraph/issues/16218), [#pull/27065](https://github.com/sourcegraph/sourcegraph/pull/27065)
- Blur omnibox/src after search [#issues/25000](https://github.com/sourcegraph/sourcegraph/issues/25000), [#pull/25004](https://github.com/sourcegraph/sourcegraph/pull/25004)
- Fix options popup toggle button styles [#pull/25210](https://github.com/sourcegraph/sourcegraph/pull/25210)
- Fix hover for files with braces [#issues/25156](https://github.com/sourcegraph/sourcegraph/issues/25156) [#pull/26278](https://github.com/sourcegraph/sourcegraph/pull/26278)
- Bitbucket cloud support [#25084](https://github.com/sourcegraph/sourcegraph/pull/25084)
