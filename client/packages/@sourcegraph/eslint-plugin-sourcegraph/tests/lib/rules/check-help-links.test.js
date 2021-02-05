'use strict'

//------------------------------------------------------------------------------
// Requirements
//------------------------------------------------------------------------------

var rule = require('../../../lib/rules/check-help-links'),
  RuleTester = require('eslint').RuleTester

// Set up the configuration such that JSX is valid.
RuleTester.setDefaultConfig({
  parserOptions: {
    ecmaVersion: 6,
    ecmaFeatures: {
      jsx: true,
    },
  },
})

//------------------------------------------------------------------------------
// Tests
//------------------------------------------------------------------------------

const ruleTester = new RuleTester()
const invalidLinkError = path => {
  return { message: 'Help link to non-existent page: ' + path, type: 'JSXOpeningElement' }
}
const options = [{ docsiteList: ['a.md', 'b/c.md', 'd/index.md'] }]

// Build up the test cases given the various combinations we need to support.
const cases = { valid: [], invalid: [] }
for (const [element, attribute] of [
  ['a', 'href'],
  ['Link', 'to'],
]) {
  for (const anchor of ['', '#anchor', '#anchor#double']) {
    for (const content of ['', 'link content']) {
      const code = target =>
        content
          ? `<${element} ${attribute}="${target}${anchor}">${content}</${element}>`
          : `<${element} ${attribute}="${target}${anchor}" />`

      cases.valid.push(
        ...[
          '/help/a',
          '/help/b/c',
          '/help/d',
          '/help/d/',
          'not-a-help-link',
          'help/but-not-absolute',
          '/help-but-not-a-directory',
        ].map(target => {
          return {
            code: code(target),
            options,
          }
        })
      )

      cases.invalid.push(
        ...['/help/', '/help/b', '/help/does/not/exist'].map(target => {
          return {
            code: code(target),
            errors: [invalidLinkError(target.substring(6))],
            options,
          }
        })
      )
    }
  }
}

// Every case should be valid if the options are empty.
cases.valid.push(
  ...[...cases.invalid, ...cases.valid].map(({ code }) => {
    return { code }
  })
)

// Actually run the tests.
ruleTester.run('check-help-links', rule, cases)
