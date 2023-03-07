const IS_BAZEL = !!(process.env.JS_BINARY__TARGET || process.env.BAZEL_BINDIR || process.env.BAZEL_TEST)
const rootDir = IS_BAZEL ? process.cwd() : __dirname

module.exports = {
  require: [
    ...(IS_BAZEL ? [] : ['ts-node/register/transpile-only']),
    'abort-controller/polyfill',
    rootDir + '/client/shared/dev/fetch',
    rootDir + '/client/shared/dev/suppressPollyErrors',
  ],
  reporter: rootDir + '/client/shared/dev/customMochaSpecReporter.js',
  extension: IS_BAZEL ? ['js'] : ['js', 'ts'],
  // 1 minute test timeout. This must be greater than the default Puppeteer
  // command timeout of 30s in order to get the stack trace to point to the
  // Puppeteer command that failed instead of a cryptic test timeout
  // location.
  timeout: '60s',
  slow: '2s',
}
