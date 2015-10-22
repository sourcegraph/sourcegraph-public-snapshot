ESLint-plugin-React
===================

[![Maintenance Status][status-image]][status-url] [![NPM version][npm-image]][npm-url] [![Build Status][travis-image]][travis-url] [![Dependency Status][deps-image]][deps-url] [![Coverage Status][coverage-image]][coverage-url] [![Code Climate][climate-image]][climate-url]

React specific linting rules for ESLint

# Installation

Install [ESLint](https://www.github.com/eslint/eslint) either locally or globally.

```sh
$ npm install eslint
```

If you installed `ESLint` globally, you have to install React plugin globally too. Otherwise, install it locally.

```sh
$ npm install eslint-plugin-react
```

# Configuration

Add `plugins` section and specify ESLint-plugin-React as a plugin.

```json
{
  "plugins": [
    "react"
  ]
}
```

If it is not already the case you must also configure `ESLint` to support JSX.

```json
{
  "ecmaFeatures": {
    "jsx": true
  }
}
```

Finally, enable all of the rules that you would like to use.

```json
{
  "rules": {
    "react/display-name": 1,
    "react/forbid-prop-types": 1,
    "react/jsx-boolean-value": 1,
    "react/jsx-closing-bracket-location": 1,
    "react/jsx-curly-spacing": 1,
    "react/jsx-indent-props": 1,
    "react/jsx-max-props-per-line": 1,
    "react/jsx-no-duplicate-props": 1,
    "react/jsx-no-literals": 1,
    "react/jsx-no-undef": 1,
    "react/jsx-quotes": 1,
    "react/jsx-sort-prop-types": 1,
    "react/jsx-sort-props": 1,
    "react/jsx-uses-react": 1,
    "react/jsx-uses-vars": 1,
    "react/no-danger": 1,
    "react/no-did-mount-set-state": 1,
    "react/no-did-update-set-state": 1,
    "react/no-direct-mutation-state": 1,
    "react/no-multi-comp": 1,
    "react/no-set-state": 1,
    "react/no-unknown-property": 1,
    "react/prefer-es6-class": 1,
    "react/prop-types": 1,
    "react/react-in-jsx-scope": 1,
    "react/require-extension": 1,
    "react/self-closing-comp": 1,
    "react/sort-comp": 1,
    "react/wrap-multilines": 1
  }
}
```

# List of supported rules

* [display-name](docs/rules/display-name.md): Prevent missing `displayName` in a React component definition
* [forbid-prop-types](docs/rules/forbid-prop-types.md): Forbid certain propTypes
* [jsx-boolean-value](docs/rules/jsx-boolean-value.md): Enforce boolean attributes notation in JSX
* [jsx-closing-bracket-location](docs/rules/jsx-closing-bracket-location.md): Validate closing bracket location in JSX
* [jsx-curly-spacing](docs/rules/jsx-curly-spacing.md): Enforce or disallow spaces inside of curly braces in JSX attributes
* [jsx-indent-props](docs/rules/jsx-indent-props.md): Validate props indentation in JSX
* [jsx-max-props-per-line](docs/rules/jsx-max-props-per-line.md): Limit maximum of props on a single line in JSX
* [jsx-no-duplicate-props](docs/rules/jsx-no-duplicate-props.md): Prevent duplicate props in JSX
* [jsx-no-literals](docs/rules/jsx-no-literals.md): Prevent usage of unwrapped JSX strings
* [jsx-no-undef](docs/rules/jsx-no-undef.md): Disallow undeclared variables in JSX
* [jsx-quotes](docs/rules/jsx-quotes.md): Enforce quote style for JSX attributes
* [jsx-sort-prop-types](docs/rules/jsx-sort-prop-types.md): Enforce propTypes declarations alphabetical sorting
* [jsx-sort-props](docs/rules/jsx-sort-props.md): Enforce props alphabetical sorting
* [jsx-uses-react](docs/rules/jsx-uses-react.md): Prevent React to be incorrectly marked as unused
* [jsx-uses-vars](docs/rules/jsx-uses-vars.md): Prevent variables used in JSX to be incorrectly marked as unused
* [no-danger](docs/rules/no-danger.md): Prevent usage of dangerous JSX properties
* [no-did-mount-set-state](docs/rules/no-did-mount-set-state.md): Prevent usage of `setState` in `componentDidMount`
* [no-did-update-set-state](docs/rules/no-did-update-set-state.md): Prevent usage of `setState` in `componentDidUpdate`
* [no-direct-mutation-state](docs/rules/no-direct-mutation-state.md): Prevent direct mutation of `this.state`
* [no-multi-comp](docs/rules/no-multi-comp.md): Prevent multiple component definition per file
* [no-set-state](docs/rules/no-set-state.md): Prevent usage of `setState`
* [no-unknown-property](docs/rules/no-unknown-property.md): Prevent usage of unknown DOM property
* [prefer-es6-class](docs/rules/prefer-es6-class.md): Prefer es6 class instead of createClass for React Components
* [prop-types](docs/rules/prop-types.md): Prevent missing props validation in a React component definition
* [react-in-jsx-scope](docs/rules/react-in-jsx-scope.md): Prevent missing `React` when using JSX
* [require-extension](docs/rules/require-extension.md): Restrict file extensions that may be required
* [self-closing-comp](docs/rules/self-closing-comp.md): Prevent extra closing tags for components without children
* [sort-comp](docs/rules/sort-comp.md): Enforce component methods order
* [wrap-multilines](docs/rules/wrap-multilines.md): Prevent missing parentheses around multilines JSX

## To Do

* no-deprecated: Prevent usage of deprecated methods ([React 0.12 Updated API](http://facebook.github.io/react/blog/2014/10/28/react-v0.12.html#new-terminology-amp-updated-apis))
* no-classic: Prevent usage of "classic" methods ([#2700](https://github.com/facebook/react/pull/2700))
* [Implement relevant rules from David Chang's React Style Guide](https://reactjsnews.com/react-style-guide-patterns-i-like)
* [Implement relevant rules from John Cobb's best practices and conventions](http://web-design-weekly.com/2015/01/29/opinionated-guide-react-js-best-practices-conventions/)
* [Implement relevant rules from Alexander Early's tips and best practices](http://aeflash.com/2015-02/react-tips-and-best-practices.html)

[Any rule idea is welcome !](https://github.com/yannickcr/eslint-plugin-react/issues)

# License

ESLint-plugin-React is licensed under the [MIT License](http://www.opensource.org/licenses/mit-license.php).


[npm-url]: https://npmjs.org/package/eslint-plugin-react
[npm-image]: http://img.shields.io/npm/v/eslint-plugin-react.svg?style=flat-square

[travis-url]: https://travis-ci.org/yannickcr/eslint-plugin-react
[travis-image]: http://img.shields.io/travis/yannickcr/eslint-plugin-react/master.svg?style=flat-square

[deps-url]: https://david-dm.org/yannickcr/eslint-plugin-react
[deps-image]: https://img.shields.io/david/dev/yannickcr/eslint-plugin-react.svg?style=flat-square

[coverage-url]: https://coveralls.io/r/yannickcr/eslint-plugin-react?branch=master
[coverage-image]: http://img.shields.io/coveralls/yannickcr/eslint-plugin-react/master.svg?style=flat-square

[climate-url]: https://codeclimate.com/github/yannickcr/eslint-plugin-react
[climate-image]: http://img.shields.io/codeclimate/github/yannickcr/eslint-plugin-react.svg?style=flat-square

[status-url]: https://github.com/yannickcr/eslint-plugin-react/pulse
[status-image]: http://img.shields.io/badge/status-maintained-brightgreen.svg?style=flat-square
