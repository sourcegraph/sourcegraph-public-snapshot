const { Console } = require('console')
const fs = require('fs')

const mocha = require('mocha')

const originalConsoleLog = mocha.reporters.Base.consoleLog

// SpecFileReporter is a custom Mocha reporter (https://mochajs.org/api/reporters_base.js.html) which behaves
// almost exactly like the stock standard mocha.reporters.Spec reporter, but it also writes the final report out
// to a file `./annotations/mocha-test-output-placeholder`.
//
// Note that this Reporter is *only* really useful in Buildkite itself, but the implementation can be adapted to
// other usecases such as when you want other Mocha reporters to output their results to disk.
//
// For more information see:
// - https://mochajs.org/api/reporters_base.js.html
// - https://mochajs.org/api/tutorial-custom-reporter.html
class SpecFileReporter extends mocha.reporters.Spec {
  // TODO(burmudar): Allow one to specify a filename and target directory ?
  constructor(runner, options) {
    super(runner, options)
    this.title = 'placeholder'
    this.buildkite = false

    if ('BUILDKITE' in process.env) {
      this.buildkite = true
    } else {
      console.info('Not in BUILDKITE. No annotation will be generated in ./annotations')
    }

    if ('BUILDKITE_LABEL' in process.env) {
      this.title = process.env.BUILDKIATE_LABEL
    }

    if (this.buildkite === true && typeof process.env.BUILDKITE_LABEL === undefined) {
      console.warn(
        `In Buildkite but BUILDKITE_LABEL not found in environment. Using title '${this.title || 'placeholder'}'`
      )
    }
  }

  epilogue() {
    // We first let mocha.reporters.Spec do it's usual reporting using the default console defined on Base
    // Which means the report will be written to the terminal
    super.epilogue()

    // We only output the epilogue to a file when we're in BUILDKITE and there are failures
    if (this.buildkite === true && this.failures.length > 0) {
      const customConsole = new Console({
        stdout: fs.createWriteStream(`./annotations/mocha-test-output-${this.title || 'placeholder'}`),
      })
      // We now want the Spec reporter (aka epilogue) to be written to a file, but Spec uses the console defined on Base!
      // So we swap out the consoleLog defined on Base with our customLog one
      // https://sourcegraph.com/github.com/mochajs/mocha/-/blob/lib/reporters/base.js?L43:5

      mocha.reporters.Base.consoleLog = customConsole.log
      // Generate report using custom logger
      // https://mochajs.org/api/reporters_base.js.html#line367
      super.epilogue()
      // The report has been written to a file, so now we swap the consoleLog back to the originalConsole logger

      mocha.reporters.Base.consoleLog = originalConsoleLog
    }
  }
}

module.exports = SpecFileReporter
