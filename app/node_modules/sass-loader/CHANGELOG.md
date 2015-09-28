Changelog
---------

### 1.0.3

- Fix importing css files from scss/sass [#101](https://github.com/jtangelder/sass-loader/issues/101)
- Fix importing SASS partials from includePath [#98](https://github.com/jtangelder/sass-loader/issues/98) [#110](https://github.com/jtangelder/sass-loader/issues/110)

### 1.0.2

- Fix a bug where files could not be imported across language styles [#73](https://github.com/jtangelder/sass-loader/issues/73)
- Update peer-dependency `node-sass` to `3.1.0`

### 1.0.1

- Fix SASS partials not being resolved anymore [#68](https://github.com/jtangelder/sass-loader/issues/68)
- Update peer-dependency `node-sass` to `3.0.0-beta.4`

### 1.0.0

- Moved `node-sass^3.0.0-alpha.0` to `peerDependencies` [#28](https://github.com/jtangelder/sass-loader/issues/28)
- Using webpack's module resolver as custom importer [#39](https://github.com/jtangelder/sass-loader/issues/31)
- Add synchronous compilation support for usage with [enhanced-require](https://github.com/webpack/enhanced-require) [#39](https://github.com/jtangelder/sass-loader/pull/39)
