/* eslint-disable @typescript-eslint/no-unsafe-return, @typescript-eslint/restrict-plus-operands, @typescript-eslint/no-unsafe-assignment, @typescript-eslint/restrict-template-expressions, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call */
const chalk = require('chalk')

module.exports = function printSuccessBanner(lines, log = console.log.bind(console)) {
  const lineLength = getLineLength(lines)
  const banner = '='.repeat(lineLength)
  const emptyLine = ' '.repeat(lineLength)

  log(chalk.bgGreenBright.black(banner))
  log(chalk.bgGreenBright.black(emptyLine))
  for (const line of lines) {
    log(chalk.bgGreenBright.black(padLine(line, lineLength)))
    log(chalk.bgGreenBright.black(emptyLine))
  }
  log(chalk.bgGreenBright.black(banner))
}

function padLine(content, length) {
  const spaceRequired = length - content.length
  const half = spaceRequired / 2
  let line = `${' '.repeat(half)}${content}${' '.repeat(half)}`
  if (line.length < length) {
    line += ' '
  }
  return line
}

function getLineLength(lines) {
  const longestLine = lines.reduce((accumulator, current) => {
    if (accumulator < current.length) {
      return current.length
    }
    return accumulator
  }, 0)

  return Math.max(40, longestLine + 10)
}
