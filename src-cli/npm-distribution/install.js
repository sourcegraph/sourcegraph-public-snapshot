#!/usr/bin/env node

const fs = require('fs')
const tar = require('tar')
const zlib = require('zlib')
const http = require('http')
const https = require('https')
const packageJSON = require('./package.json')

// Determine the URL of the file.
const platformName = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
}[process.platform]

let archName = {
  x64: 'amd64',
  x86: 'amd64',
  ia32: 'amd64',
  amd64: 'amd64',
  arm64: 'arm64',
}[process.arch]

if (!platformName || !archName) {
  console.error(`Cannot install src for platform ${process.platform}, architecture ${process.arch}`)
  process.exit(1)
}

const version = packageJSON.version
const assetURL = `https://github.com/sourcegraph/src-cli/releases/download/${version}/src-cli_${version}_${platformName}_${archName}.tar.gz`

// Remove previously-downloaded files.
const executableName = process.platform === 'win32' ? 'src.exe' : 'src'
if (fs.existsSync(executableName)) {
  fs.unlinkSync(executableName)
}

// Download the compressed file.
console.log(`Downloading ${assetURL}`)
get(assetURL, response => {
  if (response.statusCode > 299) {
    console.error(
      [
        'Download failed',
        '',
        `url: ${assetURL}`,
        `status: ${response.statusCode}`,
        `headers: ${JSON.stringify(response.headers, null, 2)}`,
        '',
      ].join('\n')
    )
    process.exit(1)
  }
  response
    .pipe(zlib.createGunzip())
    .pipe(tar.x({ filter: x => x === 'src' }))
    .addListener('close', () => fs.chmodSync(executableName, '755'))
})

// Follow redirects.
function get(url, callback) {
  const requestUrl = new URL(url)
  let request = https
  let requestConfig = requestUrl
  const proxyEnv = process.env['HTTPS_PROXY'] || process.env['https_proxy']

  if (proxyEnv) {
    const proxyUrl = new URL(proxyEnv)
    request = proxyUrl.protocol === 'https:' ? https : http
    requestConfig = {
      hostname: proxyUrl.hostname,
      port: proxyUrl.port,
      path: requestUrl.toString(),
      headers: {
        Host: requestUrl.hostname,
      },
    }
  }

  request.get(requestConfig, response => {
    if (response.statusCode === 301 || response.statusCode === 302) {
      get(response.headers.location, callback)
    } else {
      callback(response)
    }
  })
}
