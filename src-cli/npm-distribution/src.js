#!/usr/bin/env node

const path = require('path')
const spawn = require('child_process').spawn
const executable = process.platform === 'win32' ? 'src.exe' : 'src'
spawn(path.join(__dirname, executable), process.argv.slice(2), { stdio: 'inherit' }).on('close', process.exit)
