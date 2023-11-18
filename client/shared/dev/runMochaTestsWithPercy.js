// @ts-check

const path = require('path')
const { readFileSync } = require('fs')
const { execFileSync } = require('child_process')

const puppeteer = require('puppeteer')
const resolveBin = require('resolve-bin')

// Reads environment variables set by Bazel.
// It also adds a custom environment variable, "PERCY_BROWSER_EXECUTABLE", which points
// to the Puppeteer browser executable downloaded in the postinstall script.
/** @returns {Record<string, string>} env vars */
function getEnvVars() {
  // JS_BINARY__EXECROOT – Set by Bazel `js_run_binary` rule.
  // BAZEL_VOLATILE_STATUS_FILE – Set by Bazel when the `stamp` attribute on the `js_run_binary_rule` equals 1.
  // https://docs.aspect.build/rules/aspect_rules_js/docs/js_run_binary#stamp
  const { JS_BINARY__EXECROOT, BAZEL_VOLATILE_STATUS_FILE } = process.env

  if (!JS_BINARY__EXECROOT || !BAZEL_VOLATILE_STATUS_FILE) {
    throw new Error('Missing required environment variables')
  }

  // Read the Bazel status file and convert its contents to an object.
  // Here we provide information about the current BRANCH and COMMIT to allow Percy to
  // build the correct visual diff report and auto-accept change on `main` if we're on it.
  // https://github.com/percy/cli/blob/059ec21653a07105e223aa5a3ec1f815a7123ad7/packages/env/src/environment.js#L138-L139
  // https://bazel.build/docs/user-manual#workspace-status
  const statusFilePath = path.join(JS_BINARY__EXECROOT, BAZEL_VOLATILE_STATUS_FILE)
  const volatileEnvVariables = Object.fromEntries(
    readFileSync(statusFilePath, 'utf8')
      .split('\n')
      .filter(Boolean)
      .map(item => item.split(' '))
  )

  // Merge the custom "PERCY_BROWSER_EXECUTABLE" variable with the Bazel-provided variables
  // This is required to skip the "download Chromium" step in Percy's "exec" command.
  // https://docs.percy.io/docs/skipping-asset-discovery-browser-download#using-an-environment-variable
  const customEnvVariables = {
    ...volatileEnvVariables,
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    PERCY_BROWSER_EXECUTABLE: puppeteer.executablePath(),
  }

  // Convert the merged environment variables to a string with the "KEY=VALUE" format
  return customEnvVariables
}

// Resolve the binary paths for Percy and Mocha
const percyBin = resolveBin.sync('@percy/cli', { executable: 'percy' })
const mochaBin = resolveBin.sync('mocha')

// Extract command-line arguments to pass to Mocha
const args = process.argv.slice(2)

// Execute the final command, inheriting the stdio settings from the parent process and and wrapping
// the Mocha command with Percy's "exec" command (https://docs.percy.io/docs/cli-exec).
execFileSync(percyBin, ['exec', '--', mochaBin, ...args], {
  env: { ...process.env, ...getEnvVars() },
  stdio: 'inherit',
})
