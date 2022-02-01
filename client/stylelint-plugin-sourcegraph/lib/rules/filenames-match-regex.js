// @ts-check

const stylelint = require('stylelint')
const path = require('path')

const ruleName = 'filenames/match-regex'
const messages = stylelint.utils.ruleMessages(ruleName, {
  expected: 'Expected ...',
})

module.exports = stylelint.createPlugin(ruleName, function (primaryOption, secondaryOptionObject) {
  // @ts-ignore
  const regexp = new RegExp(secondaryOptionObject)

  return function (postcssRoot, postcssResult) {
    const filename = postcssRoot.source.input.from
    const name = path.basename(filename)
    const matchRegex = regexp.test(name)

    if (!matchRegex) {
      stylelint.utils.report({
        message: `Filename ${name} does not match the regular expression.`,
        ruleName,
        node: postcssRoot.first,
        result: postcssResult,
        line: 1,
      })
    }
  }
})

module.exports.messages = messages
