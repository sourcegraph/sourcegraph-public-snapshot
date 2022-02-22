const JSDOMEnvironment = require('jest-environment-jsdom')

module.exports = class JSDOMEnvironmentGlobal extends JSDOMEnvironment {
  constructor(config, options) {
    super(config, options)

    const global = (this.global = this.dom.window.document.defaultView)

    // JSDOM does not have SVGAElement implemented. Use a quick and dirty polyfill.
    // This does not implement href and target, which is impossible without mofifying JSDOM.
    global.SVGAElement = class SVGAElement extends global.SVGGraphicsElement {}

    // jsdom doesn't support document.queryCommandSupported(), needed for monaco-editor.
    // https://github.com/testing-library/react-testing-library/issues/546
    // eslint-disable-next-line @typescript-eslint/unbound-method
    this.dom.window.document.queryCommandSupported = () => false

    this.global.jsdom = this.dom
  }

  teardown() {
    this.global.jsdom = null

    return super.teardown()
  }
}
