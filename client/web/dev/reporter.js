const mocha = require('mocha')
const { Console } = require('console')
const fs = require('fs')
const originalConsole = mocha.reporters.Base.consoleLog

class SpecFileReporter extends mocha.reporters.Spec {
  constructor(runner, options) {
    super(runner, options)
    console.log('🚓 🚒 🧨 REPORTER')
    this.title = 'placeholder'
    this.buildkite = false

    if ('BUILDKITE' in process.env) {
      this.buildkite = true
    }

    if ('BUILDKITE_LABEL' in process.env) {
      this.title = process.env.BUILDKIATE_LABEL
    }
    console.log(`🕵 🔎 ➡️ buildkite=${this.buildkite}`)
  }

  epilogue() {
    super.epilogue()

    if (this.buildkite) {
      this.console = new Console({
        stdout: fs.createWriteStream(`./annotations/mocha-test-output-${this.title}`),
      })
      let tmp = Base.consoleLog

      let log = this.console.log
      Base.consoleLog = log
      super.epilogue()
      Base.consoleLog = tmp
    }
  }
}

module.exports = SpecFileReporter
