## [Unreleased]

## [0.7.2] - 2016-02-05

### Fixed
- `URI`: correctly store reference to value in constructor and return it when stringifying

### Removed
- Backed out rejection tracking for React Native `Promise` implementation. That code now lives in React Native.

## [0.7.1] - 2016-02-02

### Fixed

- Corrected require path issue for native `Promise` module

## [0.7.0] - 2016-01-27

### Added
- `Promise` for React Native with rejection tracking in `__DEV__` and a `finally` method
- `_shouldPolyfillES6Collection`: check if ES6 Collections need to be polyfilled.

### Removed
- `toArray`: removed in favor of using `Array.from` directly.

### Changed
- `ErrorUtils`: Re-uses any global instance that already exists
- `fetch`: Switched to `isomorphic-fetch` when a global implementation is missing
- `shallowEqual`: handles `NaN` values appropriately (as equal), now using `Object.is` semantics

## [0.6.1] - 2016-01-06

### Changed
- `getActiveElement`: no longer throws in non-browser environment (again)

## [0.6.0] - 2015-12-29

### Changed
- Flow: Original source files in `fbjs/flow/include` have been removed in favor of placing original files alongside compiled files in lib with a `.flow` suffix. This requires Flow version 0.19 or greater and a change to `.flowconfig` files to remove the include path.

## [0.5.1] - 2015-12-13

### Added
- `base62` module

## [0.5.0] - 2015-12-04

### Changed

- `getActiveElement`: No longer handles a non-existent `document`

## [0.4.0] - 2015-10-16

### Changed

- `invariant`: Message is no longer prefixed with "Invariant Violation: ".

## [0.3.2] - 2015-10-12

### Added
- Apply appropriate transform (`loose-envify`) when bundling with `browserify`

## [0.3.1] - 2015-10-01

### Fixed
- Ensure the build completes correctly before packaging

## [0.3.0] - 2015-10-01

### Added
- More modules: `memoizeStringOnly`, `joinClasses`
- `UserAgent`: Query information about current user agent

### Changed
- `fetchWithRetries`: Reject failure with an Error, not the response
- `getActiveElement`: no longer throws in non-browser environment
