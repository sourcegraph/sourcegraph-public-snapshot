# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).
This change log adheres to standards from [Keep a CHANGELOG](http://keepachangelog.com).

## [3.4.2] - 2015-09-18
### Fixed
* Only display the `jsx-quotes` deprecation warning with the default formatter ([#221][])

[3.4.2]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.4.1...v3.4.2
[#221]: https://github.com/yannickcr/eslint-plugin-react/issues/221

## [3.4.1] - 2015-09-17
### Fixed
* Fix `jsx-quotes` rule deprecation message ([#220][])

[3.4.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.4.0...v3.4.1
[#220]: https://github.com/yannickcr/eslint-plugin-react/issues/220

## [3.4.0] - 2015-09-16
### Added
* Add namespaced JSX support to `jsx-no-undef` ([#219][] @zertosh)
* Add option to `jsx-closing-bracket-location` to configure different styles for self-closing and non-empty tags ([#208][] @evocateur)

### Deprecated
* Deprecate `jsx-quotes` rule, will now trigger a warning if used ([#217][])

[3.4.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.3.2...v3.4.0
[#219]: https://github.com/yannickcr/eslint-plugin-react/pull/219
[#208]: https://github.com/yannickcr/eslint-plugin-react/pull/208
[#217]: https://github.com/yannickcr/eslint-plugin-react/issues/217

## [3.3.2] - 2015-09-10
### Changed
* Add `state` in lifecycle methods for `sort-comp` rule ([#197][] @mathieudutour)
* Treat component with render which returns `createElement` as valid ([#206][] @epmatsw)

### Fixed
* Fix allowed methods on arrayOf in `prop-types` ([#146][])
* Fix default configuration for `jsx-boolean-value` ([#210][])

[3.3.2]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.3.1...v3.3.2
[#146]: https://github.com/yannickcr/eslint-plugin-react/issues/146
[#197]: https://github.com/yannickcr/eslint-plugin-react/pull/197
[#206]: https://github.com/yannickcr/eslint-plugin-react/pull/206
[#210]: https://github.com/yannickcr/eslint-plugin-react/issues/210

## [3.3.1] - 2015-09-01
### Changed
* Update dependencies
* Update changelog to follow the Keep a CHANGELOG standards
* Documentation improvements ([#198][] @lencioni)

### Fixed
* Fix `jsx-closing-bracket-location` for multiline props ([#199][])

[3.3.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.3.0...v3.3.1
[#198]: https://github.com/yannickcr/eslint-plugin-react/pull/198
[#199]: https://github.com/yannickcr/eslint-plugin-react/issues/199

## [3.3.0] - 2015-08-26
### Added
* Add `jsx-indent-props` rule ([#15][], [#181][])
* Add `no-set-state rule` ([#197][] @markdalgleish)
* Add `jsx-closing-bracket-location` rule ([#14][], [#64][])

### Changed
* Update dependencies

### Fixed
* Fix crash on propTypes declarations with an empty body ([#193][] @mattyod)

[3.3.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.2.3...v3.3.0
[#15]: https://github.com/yannickcr/eslint-plugin-react/issues/15
[#181]: https://github.com/yannickcr/eslint-plugin-react/issues/181
[#197]: https://github.com/yannickcr/eslint-plugin-react/pull/197
[#14]: https://github.com/yannickcr/eslint-plugin-react/issues/14
[#64]: https://github.com/yannickcr/eslint-plugin-react/issues/64
[#193]: https://github.com/yannickcr/eslint-plugin-react/pull/193

## [3.2.3] - 2015-08-16
### Changed
* Update dependencies

### Fixed
* Fix object rest/spread handling ([#187][] @xjamundx, [#189][] @Morantron)

[3.2.3]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.2.2...v3.2.3
[#187]: https://github.com/yannickcr/eslint-plugin-react/pull/187
[#189]: https://github.com/yannickcr/eslint-plugin-react/pull/189

## [3.2.2] - 2015-08-11
### Changed
* Remove peerDependencies ([#178][])

[3.2.2]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.2.1...v3.2.2
[#178]: https://github.com/yannickcr/eslint-plugin-react/issues/178

## [3.2.1] - 2015-08-08
### Fixed
* Fix crash when propTypes don't have any parent ([#182][])
* Fix jsx-no-literals reporting errors outside JSX ([#183][] @CalebMorris)

[3.2.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.2.0...v3.2.1
[#182]: https://github.com/yannickcr/eslint-plugin-react/issues/182
[#183]: https://github.com/yannickcr/eslint-plugin-react/pull/183

## [3.2.0] - 2015-08-04
### Added
* Add `jsx-max-props-per-line` rule ([#13][])
* Add `jsx-no-literals` rule ([#176][] @CalebMorris)

### Changed
* Update dependencies

### Fixed
* Fix object access in `jsx-no-undef` ([#172][])

[3.2.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.1.0...v3.2.0
[#13]: https://github.com/yannickcr/eslint-plugin-react/issues/13
[#176]: https://github.com/yannickcr/eslint-plugin-react/pull/176
[#172]: https://github.com/yannickcr/eslint-plugin-react/issues/172

## [3.1.0] - 2015-07-28
### Added
* Add event handlers to `no-unknown-property` ([#164][] @mkenyon)
* Add customValidators option to `prop-types` ([#145][] @CalebMorris)

### Changed
* Update dependencies
* Documentation improvements ([#167][] @ngbrown)

### Fixed
* Fix comment handling in `jsx-curly-spacing` ([#165][])

[3.1.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v3.0.0...v3.1.0
[#164]: https://github.com/yannickcr/eslint-plugin-react/pull/164
[#145]: https://github.com/yannickcr/eslint-plugin-react/issues/145
[#165]: https://github.com/yannickcr/eslint-plugin-react/issues/165
[#167]: https://github.com/yannickcr/eslint-plugin-react/pull/167

## [3.0.0] - 2015-07-21
### Added
* Add jsx-no-duplicate-props rule ([#161][] @hummlas)
* Add allowMultiline option to the `jsx-curly-spacing` rule ([#156][] @mathieumg)

## Breaking
* In `jsx-curly-spacing` braces spanning multiple lines are now allowed with `never` option ([#156][] @mathieumg)

### Fixed
* Fix multiple var and destructuring handling in `props-types` ([#159][])
* Fix crash when retrieving propType name ([#163][])

[3.0.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.7.1...v3.0.0
[#161]: https://github.com/yannickcr/eslint-plugin-react/pull/161
[#156]: https://github.com/yannickcr/eslint-plugin-react/pull/156
[#159]: https://github.com/yannickcr/eslint-plugin-react/issues/159
[#163]: https://github.com/yannickcr/eslint-plugin-react/issues/163

## [2.7.1] - 2015-07-16
### Changed
* Update peerDependencies requirements ([#154][])
* Update codebase for ESLint v1.0.0
* Change oneOfType to actually keep the child types ([#148][] @CalebMorris)
* Documentation improvements ([#147][] @lencioni)

[2.7.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.7.0...v2.7.1
[#154]: https://github.com/yannickcr/eslint-plugin-react/issues/154
[#148]: https://github.com/yannickcr/eslint-plugin-react/issues/148
[#147]: https://github.com/yannickcr/eslint-plugin-react/pull/147

## [2.7.0] - 2015-07-11
### Added
* Add `no-danger` rule ([#138][] @scothis)
* Add `jsx-curly-spacing` rule ([#142][])

### Fixed
* Fix properties limitations on propTypes ([#139][])
* Fix component detection ([#144][])

[2.7.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.6.4...v2.7.0
[#138]: https://github.com/yannickcr/eslint-plugin-react/pull/138
[#142]: https://github.com/yannickcr/eslint-plugin-react/issues/142
[#139]: https://github.com/yannickcr/eslint-plugin-react/issues/139
[#144]: https://github.com/yannickcr/eslint-plugin-react/issues/144

## [2.6.4] - 2015-07-02
### Fixed
* Fix simple destructuring handling ([#137][])

[2.6.4]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.6.3...v2.6.4
[#137]: https://github.com/yannickcr/eslint-plugin-react/issues/137

## [2.6.3] - 2015-06-30
### Fixed
* Fix ignore option for `prop-types` rule ([#135][])
* Fix nested props destructuring ([#136][])

[2.6.3]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.6.2...v2.6.3
[#135]: https://github.com/yannickcr/eslint-plugin-react/issues/135
[#136]: https://github.com/yannickcr/eslint-plugin-react/issues/136

## [2.6.2] - 2015-06-28
### Fixed
* Fix props validation when using a prop as an object key ([#132][])

[2.6.2]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.6.1...v2.6.2
[#132]: https://github.com/yannickcr/eslint-plugin-react/issues/132

## [2.6.1] - 2015-06-28
### Fixed
* Fix crash in `prop-types` when encountering an empty variable declaration ([#130][])

[2.6.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.6.0...v2.6.1
[#130]: https://github.com/yannickcr/eslint-plugin-react/issues/130

## [2.6.0] - 2015-06-28
### Added
* Add support for nested propTypes ([#62][] [#105][] @Cellule)
* Add `require-extension` rule ([#117][] @scothis)
* Add support for computed string format in `prop-types` ([#127][] @Cellule)
* Add ES6 methods to `sort-comp` default configuration ([#97][] [#122][])
* Add support for props destructuring directly on the this keyword
* Add `acceptTranspilerName` option to `display-name` rule ([#75][])
* Add schema to validate rules options

### Changed
* Update dependencies

### Fixed
* Fix test command for Windows ([#114][] @Cellule)
* Fix detection of missing displayName and propTypes when `ecmaFeatures.jsx` is false ([#119][] @rpl)
* Fix propTypes destructuring with properties as string ([#118][] @Cellule)
* Fix `jsx-sort-prop-types` support for keys as string ([#123][] @Cellule)
* Fix crash if a ClassProperty has only one token ([#125][])
* Fix invalid class property handling in `jsx-sort-prop-types` ([#129][])

[2.6.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.5.2...v2.6.0
[#62]: https://github.com/yannickcr/eslint-plugin-react/issues/62
[#105]: https://github.com/yannickcr/eslint-plugin-react/issues/105
[#114]: https://github.com/yannickcr/eslint-plugin-react/pull/114
[#117]: https://github.com/yannickcr/eslint-plugin-react/pull/117
[#119]: https://github.com/yannickcr/eslint-plugin-react/pull/119
[#118]: https://github.com/yannickcr/eslint-plugin-react/issues/118
[#123]: https://github.com/yannickcr/eslint-plugin-react/pull/123
[#125]: https://github.com/yannickcr/eslint-plugin-react/issues/125
[#127]: https://github.com/yannickcr/eslint-plugin-react/pull/127
[#97]: https://github.com/yannickcr/eslint-plugin-react/issues/97
[#122]: https://github.com/yannickcr/eslint-plugin-react/issues/122
[#129]: https://github.com/yannickcr/eslint-plugin-react/issues/129
[#75]: https://github.com/yannickcr/eslint-plugin-react/issues/75

## [2.5.2] - 2015-06-14
### Fixed
* Fix regression in `jsx-uses-vars` with `babel-eslint` ([#110][])

[2.5.2]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.5.1...v2.5.2
[#110]: https://github.com/yannickcr/eslint-plugin-react/issues/110

## [2.5.1] - 2015-06-14
### Changed
* Update dependencies
* Documentation improvements ([#99][] @morenoh149)

### Fixed
* Fix `prop-types` crash when propTypes definition is invalid ([#95][])
* Fix `jsx-uses-vars` for ES6 classes ([#96][])
* Fix hasOwnProperty that is taken for a prop ([#102][])

[2.5.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.5.0...v2.5.1
[#95]: https://github.com/yannickcr/eslint-plugin-react/issues/95
[#96]: https://github.com/yannickcr/eslint-plugin-react/issues/96
[#102]: https://github.com/yannickcr/eslint-plugin-react/issues/102
[#99]: https://github.com/yannickcr/eslint-plugin-react/pull/99

## [2.5.0] - 2015-06-04
### Added
* Add option to make `wrap-multilines` more granular ([#94][] @PiPeep)

### Changed
* Update dependencies
* Documentation improvements ([#92][] [#93][] @lencioni)

[2.5.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.4.0...v2.5.0
[#94]: https://github.com/yannickcr/eslint-plugin-react/pull/94
[#92]: https://github.com/yannickcr/eslint-plugin-react/pull/92
[#93]: https://github.com/yannickcr/eslint-plugin-react/pull/93

## [2.4.0] - 2015-05-30
### Added
* Add pragma option to `jsx-uses-react` ([#82][] @dominicbarnes)
* Add context props to `sort-comp` ([#89][] @zertosh)

### Changed
* Update dependencies
* Documentation improvement ([#91][] @matthewwithanm)

### Fixed
* Fix itemID in `no-unknown-property` rule ([#85][] @cody)
* Fix license field in package.json ([#90][] @zertosh)
* Fix usage of contructor in `sort-comp` options ([#88][])

[2.4.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.3.0...v2.4.0
[#82]: https://github.com/yannickcr/eslint-plugin-react/pull/82
[#89]: https://github.com/yannickcr/eslint-plugin-react/pull/89
[#85]: https://github.com/yannickcr/eslint-plugin-react/pull/85
[#90]: https://github.com/yannickcr/eslint-plugin-react/pull/90
[#88]: https://github.com/yannickcr/eslint-plugin-react/issues/88
[#91]: https://github.com/yannickcr/eslint-plugin-react/pull/91

## [2.3.0] - 2015-05-14
### Added
* Add `sort-comp` rule ([#39][])
* Add `allow-in-func` option to `no-did-mount-set-state` ([#56][])

### Changed
* Update dependencies
* Improve errors locations for `prop-types`

### Fixed
* Fix quoted propTypes in ES6 ([#77][])

[2.3.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.2.0...v2.3.0
[#39]: https://github.com/yannickcr/eslint-plugin-react/issues/39
[#77]: https://github.com/yannickcr/eslint-plugin-react/issues/77
[#56]: https://github.com/yannickcr/eslint-plugin-react/issues/56

## [2.2.0] - 2015-04-22
### Added
* Add `jsx-sort-prop-types` rule ([#38][] @AlexKVal)

### Changed
* Documentation improvements ([#71][] @AlexKVal)

### Fixed
* Fix variables marked as used when a prop has the same name ([#69][] @burnnat)

[2.2.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.1.1...v2.2.0
[#38]: https://github.com/yannickcr/eslint-plugin-react/issues/38
[#69]: https://github.com/yannickcr/eslint-plugin-react/pull/69
[#71]: https://github.com/yannickcr/eslint-plugin-react/pull/71

## [2.1.1] - 2015-04-17
### Added
* Add support for classes static properties ([#43][])
* Add tests for the `babel-eslint` parser
* Add ESLint as peerDependency ([#63][] @AlexKVal)

### Changed
* Documentation improvements ([#55][] @AlexKVal, [#60][] @chriscalo)

[2.1.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.1.0...v2.1.1
[#43]: https://github.com/yannickcr/eslint-plugin-react/issues/43
[#63]: https://github.com/yannickcr/eslint-plugin-react/pull/63
[#55]: https://github.com/yannickcr/eslint-plugin-react/pull/55
[#60]: https://github.com/yannickcr/eslint-plugin-react/pull/60

## [2.1.0] - 2015-04-06
### Added
* Add `jsx-boolean-value` rule ([#11][])
* Add support for static methods in `display-name` and `prop-types` ([#48][])

### Changed
* Update `jsx-sort-props` to reset the alphabetical verification on spread ([#47][] @zertosh)
* Update `jsx-uses-vars` to be enabled by default ([#49][] @banderson)

### Fixed
* Fix describing comment for hasSpreadOperator() method ([#53][] @AlexKVal)

[2.1.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.0.2...v2.1.0
[#47]: https://github.com/yannickcr/eslint-plugin-react/pull/47
[#49]: https://github.com/yannickcr/eslint-plugin-react/pull/49
[#11]: https://github.com/yannickcr/eslint-plugin-react/issues/11
[#48]: https://github.com/yannickcr/eslint-plugin-react/issues/48
[#53]: https://github.com/yannickcr/eslint-plugin-react/pull/53

## [2.0.2] - 2015-03-31
### Fixed
* Fix ignore rest spread when destructuring props ([#46][])
* Fix component detection in `prop-types` and `display-name` ([#45][])
* Fix spread handling in `jsx-sort-props` ([#42][] @zertosh)

[2.0.2]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.0.1...v2.0.2
[#46]: https://github.com/yannickcr/eslint-plugin-react/issues/46
[#45]: https://github.com/yannickcr/eslint-plugin-react/issues/45
[#42]: https://github.com/yannickcr/eslint-plugin-react/pull/42

## [2.0.1] - 2015-03-30
### Fixed
* Fix props detection when used in an object ([#41][])

[2.0.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v2.0.0...v2.0.1
[#41]: https://github.com/yannickcr/eslint-plugin-react/issues/41

## [2.0.0] - 2015-03-29
### Added
* Add `jsx-sort-props` rule ([#16][])
* Add `no-unknown-property` rule ([#28][])
* Add ignore option to `prop-types` rule

### Changed
* Update dependencies

## Breaking
* In `prop-types` the children prop is no longer ignored

### Fixed
* Fix components are now detected when using ES6 classes ([#24][])
* Fix `prop-types` now return the right line/column ([#33][])
* Fix props are now detected when destructuring ([#27][])
* Fix only check for computed property names in `prop-types` ([#36][] @burnnat)

[2.0.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.6.1...v2.0.0
[#16]: https://github.com/yannickcr/eslint-plugin-react/issues/16
[#28]: https://github.com/yannickcr/eslint-plugin-react/issues/28
[#24]: https://github.com/yannickcr/eslint-plugin-react/issues/24
[#33]: https://github.com/yannickcr/eslint-plugin-react/issues/33
[#27]: https://github.com/yannickcr/eslint-plugin-react/issues/27
[#36]: https://github.com/yannickcr/eslint-plugin-react/pull/36

## [1.6.1] - 2015-03-25
### Changed
* Update `jsx-quotes` documentation

### Fixed
* Fix `jsx-no-undef` with `babel-eslint` ([#30][])
* Fix `jsx-quotes` on Literal childs ([#29][])

[1.6.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.6.0...v1.6.1
[#30]: https://github.com/yannickcr/eslint-plugin-react/issues/30
[#29]: https://github.com/yannickcr/eslint-plugin-react/issues/29

## [1.6.0] - 2015-03-22
### Added
* Add `jsx-no-undef` rule
* Add `jsx-quotes` rule ([#12][]) 
* Add `@jsx` pragma support ([#23][])

### Changed
* Allow `this.getState` references (not calls) in lifecycle methods ([#22][] @benmosher)
* Update dependencies

### Fixed
* Fix `react-in-jsx-scope` in Node.js env
* Fix usage of propTypes with an external object ([#9][])

[1.6.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.5.0...v1.6.0
[#12]: https://github.com/yannickcr/eslint-plugin-react/issues/12
[#23]: https://github.com/yannickcr/eslint-plugin-react/issues/23
[#9]: https://github.com/yannickcr/eslint-plugin-react/issues/9
[#22]: https://github.com/yannickcr/eslint-plugin-react/pull/22

## [1.5.0] - 2015-03-14
### Added
* Add `jsx-uses-vars` rule

### Fixed
* Fix `jsx-uses-react` for ESLint 0.17.0

[1.5.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.4.1...v1.5.0

## [1.4.1] - 2015-03-03
### Fixed
* Fix `this.props.children` marked as missing in props validation ([#7][])
* Fix usage of `this.props` without property ([#8][])

[1.4.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.4.0...v1.4.1
[#7]: https://github.com/yannickcr/eslint-plugin-react/issues/7
[#8]: https://github.com/yannickcr/eslint-plugin-react/issues/8

## [1.4.0] - 2015-02-24
### Added
* Add `react-in-jsx-scope` rule ([#5][] @glenjamin)
* Add `jsx-uses-react` rule ([#6][] @glenjamin)

### Changed
* Update `prop-types` to check props usage insead of propTypes presence ([#4][])

[1.4.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.3.0...v1.4.0
[#4]: https://github.com/yannickcr/eslint-plugin-react/issues/4
[#5]: https://github.com/yannickcr/eslint-plugin-react/pull/5
[#6]: https://github.com/yannickcr/eslint-plugin-react/pull/6

## [1.3.0] - 2015-02-24
### Added
* Add `no-did-mount-set-state` rule
* Add `no-did-update-set-state` rule

### Changed
* Update dependencies

[1.3.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.2.2...v1.3.0

## [1.2.2] - 2015-02-09
### Changed
* Update dependencies

### Fixed
* Fix childs detection in `self-closing-comp` ([#3][])

[1.2.2]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.2.1...v1.2.2
[#3]: https://github.com/yannickcr/eslint-plugin-react/issues/3

## [1.2.1] - 2015-01-29
### Changed
* Update Readme
* Update dependencies
* Update `wrap-multilines` and `self-closing-comp` rules for ESLint 0.13.0

[1.2.1]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.2.0...v1.2.1

## [1.2.0] - 2014-12-29
### Added
* Add `self-closing-comp` rule

### Fixed
* Fix `display-name` and `prop-types` rules

[1.2.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.1.0...v1.2.0

## [1.1.0] - 2014-12-28
### Added
 * Add `display-name` rule
 * Add `wrap-multilines` rule
 * Add rules documentation
 * Add rules tests

[1.1.0]: https://github.com/yannickcr/eslint-plugin-react/compare/v1.0.0...v1.1.0

## 1.0.0 - 2014-12-16
### Added
 * First revision
