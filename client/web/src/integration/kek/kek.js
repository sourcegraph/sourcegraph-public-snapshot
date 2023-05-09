const path = require('path')
const fs = require('fs')
const resolveBin = require('resolve-bin')
const shelljs = require('shelljs')
const puppeteer = require('puppeteer')

const statusFilePath = path.join(process.env.JS_BINARY__EXECROOT, process.env.BAZEL_VOLATILE_STATUS_FILE)
const statusFile = Object.fromEntries(
  fs
    .readFileSync(statusFilePath, 'utf8')
    .split('\n')
    .map(item => item.split(' '))
)

const percyCli = resolveBin.sync('@percy/cli', { executable: 'percy' })
const mocha = resolveBin.sync('mocha')

console.log('----------------------------------------------')
console.log('executablePath', puppeteer.executablePath())
console.log('[STATUS FILE]', statusFile)
console.log('[JS_BINARY__EXECROOT]', process.env.JS_BINARY__EXECROOT)
console.log('[CWD]', process.cwd())
console.log('----------------------------------------------')

console.log('----------------------------------------------')
const args = process.argv.slice(2)
const cmd = `${percyCli} exec --verbose -- ${mocha} ${args.join(' ')}`
console.log('[CMD]', cmd)
const res = shelljs.exec(cmd)
console.log('----------------------------------------------')

if (res.code !== 0) {
  throw new Error('test failed')
}

// throw new Error(`${mocha} ${percyCli} ${JSON.stringify(statusFile)}`)
