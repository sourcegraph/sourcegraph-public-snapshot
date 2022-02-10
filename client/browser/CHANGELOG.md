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

- Make "single click to definition" an opt-in through advanced settings: [#pull/30540](https://github.com/sourcegraph/sourcegraph/pull/30540), [#issues/#30437](https://github.com/sourcegraph/sourcegraph/issues/30437)
- Add "https://" URL input placeholder [#pull/30282](https://github.com/sourcegraph/sourcegraph/pull/30282), [#issues/14723](https://github.com/sourcegraph/sourcegraph/issues/14723)
- Add filtering of browser extension dropdown duplicated URLs [#pull/30674](https://github.com/sourcegraph/sourcegraph/pull/30674), [#issues/30673](https://github.com/sourcegraph/sourcegraph/issues/30673)
- Add code intel support to GitHub pull request commit view [#pull/30618](https://github.com/sourcegraph/sourcegraph/pull/30618), [#issues/30623](https://github.com/sourcegraph/sourcegraph/issues/30623)
- Add tracking of inbound traffic from browser extension/code host integration [#pull/30170](https://github.com/sourcegraph/sourcegraph/pull/30170), [#issues/27082](https://github.com/sourcegraph/sourcegraph/issues/27082)

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
