## [Unreleased]

## 0.3.2 - 2015-01-12

### Added
- Apply appropriate transform (`loose-envify`) when bundling with `browserify`

## 0.3.1 - 2015-10-01

### Fixed
- Ensure the build completes correctly before packaging

## [0.3.0] - 2015-10-01

### Added
- More modules: `memoizeStringOnly`, `joinClasses`
- `UserAgent`: Query information about current user agent

### Changed
- `fetchWithRetries`: Reject failure with an Error, not the response
- `getActiveElement`: no longer throws in non-browser environment
