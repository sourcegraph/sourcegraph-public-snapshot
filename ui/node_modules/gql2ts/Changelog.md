## 0.4.0
- Stop extending `GraphQLInterface`s with their possible types. [#25](https://github.com/avantcredit/gql2ts/pull/25)
  - Previously, if two possible types implement a similar field, but with a different type it will cause an error

## 0.3.1
- Accept `__schema` at the top level [#20](https://github.com/avantcredit/gql2ts/pull/20)

## 0.3.0
### Breaking Changes
- Change from `module` to `namespace` [#14](https://github.com/avantcredit/gql2ts/pull/14)
- Removed `-m`/`--module-name` flag in favor of `-n`/`--namespace`
### Patches
- Fix for `Int` Scalar Type (thanks [@valorize](https://github.com/valorize)) [#15](https://github.com/avantcredit/gql2ts/pull/15)

## 0.2.1
- Fix Version number in command line

## 0.2.0
- Add support for Enums

## 0.1.0
- Add Root Entry Points & Error Map
- Add `__typename` to the generated interfaces

## 0.0.4
- Include polyfill `Array.prototype.includes` for node v4/v5 compatibility
- Add test suite

## 0.0.3
- Fix for node v5 strict mode

## 0.0.2
- Add information to npm

## 0.0.1
- Initial Release
