// @ts-check

const path = require('path')
const fs = require('fs')
const { execSync } = require('child_process')

const puppeteer = require('puppeteer')
const resolveBin = require('resolve-bin')

function getEnvVariablesString() {
  // TODO: add comment on how this variables are set by Bazel.
  const { JS_BINARY__EXECROOT, BAZEL_VOLATILE_STATUS_FILE } = process.env

  if (!JS_BINARY__EXECROOT || !BAZEL_VOLATILE_STATUS_FILE) {
    throw new Error('Missing required environment variables')
  }

  const statusFilePath = path.join(JS_BINARY__EXECROOT, BAZEL_VOLATILE_STATUS_FILE)
  const volatileEnvVariables = Object.fromEntries(
    fs
      .readFileSync(statusFilePath, 'utf8')
      .split('\n')
      .filter(Boolean)
      .map(item => item.split(' '))
  )

  const customEnvVariables = {
    ...volatileEnvVariables,
    // @ts-ignore
    PERCY_BROWSER_EXECUTABLE: puppeteer.executablePath(),
  }

  return Object.entries(customEnvVariables)
    .map(([key, value]) => `${key}=${value}`)
    .join(' ')
}

const percyBin = resolveBin.sync('@percy/cli', { executable: 'percy' })
const mochaBin = resolveBin.sync('mocha')

const args = process.argv.slice(2)
const testCmd = `${mochaBin} ${args.join(' ')}`
const finalCmd = `${getEnvVariablesString()} ${percyBin} exec -- ${testCmd}`

execSync(finalCmd, { stdio: 'inherit' })
