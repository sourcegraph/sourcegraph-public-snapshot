const JestNodeEnvironment = require('jest-environment-node')

module.exports = class NodeEnvironmentGlobal extends JestNodeEnvironment {
  constructor(config, options) {
    super(config, options)
    this.global.hasTestFailures = false
  }

  teardown() {
    this.global.hasTestFailures = false

    return super.teardown()
  }

  handleTestEvent(event) {
    if (event.name === 'test_fn_failure') {
      this.global.hasTestFailures = true
    }
  }
}
