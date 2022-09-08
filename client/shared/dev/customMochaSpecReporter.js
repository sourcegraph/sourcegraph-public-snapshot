const { execSync } = require('child_process')
const { Console } = require('console')
const fs = require('fs')

const repoRoot = execSync('git rev-parse --show-toplevel').toString().trimEnd()
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

    if ('BUILDKITE' in process.env) {
      this.buildkite = true
    } else {
      console.info('Not in BUILDKITE. No annotation will be generated in ./annotations')
    }

    if ('BUILDKITE_LABEL' in process.env) {
      this.title = process.env.BUILDKITE_LABEL || 'placeholder'
    }

    if (this.buildkite === true && typeof process.env.BUILDKITE_LABEL === undefined) {
      console.info(`In Buildkite but BUILDKITE_LABEL not found in environment. Using title '${this.title}'`)
    }
  }

  safeClose(stream) {
    return new Promise((resolve, reject) => {
      stream.close(error => {
        if (error) {
          reject(error)
        }

        resolve()
      })
    })
  }

  epilogue() {
    // We first let mocha.reporters.Spec do it's usual reporting using the default console defined on Base
    // Which means the report will be written to the terminal
    super.epilogue()

    // We only output the epilogue to a file when we're in BUILDKITE and there are failures
    if (this.buildkite === true && this.failures.length > 0) {
      const file = fs.createWriteStream(`${repoRoot}/annotations/${this.title}`)
      const customConsole = new Console({
        stdout: file,
      })

      // We now want the Spec reporter (aka epilogue) to be written to a file, but Spec uses the console defined on Base!
      // So we swap out the consoleLog defined on Base with our customLog one
      // https://sourcegraph.com/github.com/mochajs/mocha/-/blob/lib/reporters/base.js?L43:5
      // eslint-disable-next-line @typescript-eslint/unbound-method
      mocha.reporters.Base.consoleLog = customConsole.log
      // Generate report using custom logger
      // https://mochajs.org/api/reporters_base.js.html#line367
      super.epilogue()
      // The report has been written to a file, so now we swap the consoleLog back to the originalConsole logger
      mocha.reporters.Base.consoleLog = originalConsoleLog
      // We want to make sure before this reporter exits that the data written to file have been flushed
      // In some scenarios, the node process exits too quickly and the data hasn't been flushed to the file yet
      this.safeClose(file)
        .then(() => {
          const path = file.path.toString()
          console.log(`${path} successfully closed`)
        })
        .catch(error => console.error(error))
        .finally(() => {
          console.warn('force exiting the process after writing report to file')
          // This performs the same function as passing --exit to the mocha test runner
          // When the regression tests run, some resources are not properly cleaned up. Leading to
          // the test runner just hanging since it is waiting on an open resource to exit.
          // TODO(burmudar): hunt this resource down
          process.exit(1)
        })
    }
  }
}

module.exports = SpecFileReporter
