# Change Log
All notable changes to this project will be documented in this file.

## [next]
### Features

### Fixed

### Tests

## [2.0.0]
### BREAKING CHANGES
- `.sass` files are not included in the graph by default. Use the `-e .sass` flag.

### Features
- Configurable file extensions - [@dannymidnight](https://github.com/dannymidnight), [@xzyfer](https://github.com/xzyfer)

### Fixed
- Prioritize cwd when resolving load paths - [@schnerd](https://github.com/schnerd)

### Tests
- Added test for prioritizing cwd when resolving load paths - [@xzyfer](https://github.com/xzyfer)

## [1.3.0]
### Features
- Add support for indented syntax - [@vegetableman](https://github.com/vegetableman)

## [1.2.0]
### Features
- Add support for custom imports - [@kevin-smets](https://github.com/kevin-smets)

## [1.1.0] - 2015-03-18
### Fixed
- Only strip extension for css, scss, sass files - [@nervo](https://github.com/nervo)

## [1.0.4] - 2015-03-03
### Tests
- Added a test for nested imports - [@kevin-smets](https://github.com/kevin-smets)

## [1.0.3] - 2015-02-02
### Fixed
- Replace incorrect usage of `for..in` loops with simple `for` loops

## [1.0.2] - 2015-02-02
### Fixed
- Don't iterate over inherited object properties

## [1.0.1] - 2015-01-05
### Fixed
- Handle errors in the visitor

## [1.0.0] - 2015-01-05

Initial stable release
