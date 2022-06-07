# Changelog

The Sourcegraph extension uses major.EVEN_NUMBER.patch (eg. 2.0.1) for release versions and major.ODD_NUMBER.patch (eg. 2.1.1) for pre-release versions.

## Next Release

### Changes

- Import Sourcegraph Access token automatically after completing authentication in the browser for Sourcegraph Cloud users [issues/28311](https://github.com/sourcegraph/sourcegraph/issues/28311)

### Fixes

- Optimize build size [issues/36192](https://github.com/sourcegraph/sourcegraph/issues/36192)

## 2.2.3

### Changes

- Update Access Token headers setting method --thanks @ptxmac for the contribution! [issues/34338](https://github.com/sourcegraph/sourcegraph/issues/34338)
- Add options to choose between main branch or current branch when copy/open file. Always use default branch if set [issues/34591](https://github.com/sourcegraph/sourcegraph/issues/34591)
- CTA for adding Sourcegraph extension to Workspace Recommendations [issues/34829](https://github.com/sourcegraph/sourcegraph/issues/34829)

### Fixes

- Windows file path issue [issues/34788](https://github.com/sourcegraph/sourcegraph/issues/34788)
- Sourcegraph icon in help sidebar now shows on light theme [issues/35672](https://github.com/sourcegraph/sourcegraph/issues/35672)
- Highlight background color for VS Code Light & Light+ Theme [issues/35767](https://github.com/sourcegraph/sourcegraph/issues/35767)
- Display reload button when instance URL is updated [issues/35980](https://github.com/sourcegraph/sourcegraph/issues/35980)

## 2.2.2

### Changes

- Display current extension version and instance version in frontend [issues/34729](https://github.com/sourcegraph/sourcegraph/issues/34729)

### Fixes

- Remove incorrect unsupported instance error messages on first load [issues/34207](https://github.com/sourcegraph/sourcegraph/issues/34207)
- Links to open remote file in Sourcegraph web are now decoded correctly [issues/34630](https://github.com/sourcegraph/sourcegraph/issues/34630)
- Remove pattern restriction for basePath [issues/34731](https://github.com/sourcegraph/sourcegraph/issues/34731)

## 2.2.1

### Changes

- Add Help and Feedback sidebar [issue/31021](https://github.com/sourcegraph/sourcegraph/issues/31021)
- Add CONTRIBUTING guide [issue/26536](https://github.com/sourcegraph/sourcegraph/issues/26536)
- Display error message when connected to unsupported instances [issue/31808](https://github.com/sourcegraph/sourcegraph/issues/31808)
- Log events with `IDEEXTENSION` as event source for instances on 3.38.0 and above [issue/32851](https://github.com/sourcegraph/sourcegraph/issues/32851)
- Add new configuration setting: sourcegraph.basePath [issue/32633](https://github.com/sourcegraph/sourcegraph/issues/32633)
- Add ability to open local copy of a search result if file exists in current workspace or basePath [issue/32633](https://github.com/sourcegraph/sourcegraph/issues/32633)

### Fixes

- Improve developer scripts [issue/32741](https://github.com/sourcegraph/sourcegraph/issues/32741)
- Code Monitor button redirect issue for non signed-in users [issues/33631](https://github.com/sourcegraph/sourcegraph/issues/33631)
- Error regarding missing PatternType when creating save search [issues/31093](https://github.com/sourcegraph/sourcegraph/issues/31093)

## 2.2.0

### Changes

- Add pings for Sourcegraph ide extensions usage metrics [issue/29124](https://github.com/sourcegraph/sourcegraph/issues/29124)
- Add input fields to update Sourcegraph instance url [issue/31804](https://github.com/sourcegraph/sourcegraph/issues/31804)
- Clear search results on tab close [issue/30583](https://github.com/sourcegraph/sourcegraph/issues/30583)

## 2.0.9

### Changes

- Add Changelog for version tracking purpose [issue/28300](https://github.com/sourcegraph/sourcegraph/issues/28300)
- Add VS Code Web support for instances on 3.36.0+ [issue/28403](https://github.com/sourcegraph/sourcegraph/issues/28403)
- Update to use API endpoint for stream search [issue/30916](https://github.com/sourcegraph/sourcegraph/issues/30916)
- Add new configuration setting `sourcegraph.requestHeaders` for adding custom headers [issue/30916](https://github.com/sourcegraph/sourcegraph/issues/30916)

### Fixes

- Manage context display issue for instances under v3.36.0 [issue/31022](https://github.com/sourcegraph/sourcegraph/issues/31022)

## 2.0.8

### Fixes

- Files will open in the correct url scheme [issue/31095](https://github.com/sourcegraph/sourcegraph/issues/31095)
- The 'All Search Keywords' button is now linked to Sourcegraph docs site correctly [issue/31023](https://github.com/sourcegraph/sourcegraph/issues/31023)
- Update Sign Up links with the correct utm parameters

## 2.0.7

### Changes

- Remove Sign Up CTA in Sidebar for self-host instances

### Fixes

- Add backward compatibility for configuration settings from v1: `sourcegraph.defaultBranch` and `sourcegraph.remoteUrlReplacements`

## 2.0.6

### Changes

- Remove Sign Up CTAs in Search Result for self-host instances

## 2.0.1

### Changes

- Add Code Monitor
