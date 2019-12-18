// @ts-check

const { spawn } = require('child_process')
const gulp = require('gulp')
const path = require('path')

function build() {
  return spawn('yarn', ['-s', 'run', 'build'], {
    stdio: 'inherit',
    shell: true,
    env: { ...process.env, NODE_OPTIONS: '--max_old_space_size=8192' },
  })
}

function watch() {
  return spawn('yarn', ['-s', 'run', 'dev'], {
    stdio: 'inherit',
    shell: true,
    env: { ...process.env, NODE_OPTIONS: '--max_old_space_size=8192' },
  })
}

module.exports = { build, watch }
