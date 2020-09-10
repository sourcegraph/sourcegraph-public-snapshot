const path = require('path')
const shelljs = require('shelljs')
const extensionNames = require('../bundled-code-intel-extensions.json').extensions || []
const toDirectory = path.join(process.cwd(), 'build')
const temporarySourceDirectory = path.join(process.cwd(), 'code-intel-extensions')
const signale = require('signale')

// Install dependencies and build the specified code intel extension
shelljs.exec('yarn --cwd code-intel-extensions install')

// TODO: check if code-intel-extensions exists
// TODO: for each extension, check if that particular directory exists
// (and abort when yarn fails)
for (const extensionName of extensionNames) {
  shelljs.exec(`yarn --cwd ${temporarySourceDirectory}/extensions/${extensionName} run build`)
}
// shelljs.popd()

for (const extensionName of extensionNames) {
  // Copy extension manifest (package.json) and bundle (extension.js)
  shelljs.mkdir('-p', `${toDirectory}/extensions/${extensionName}`)

  shelljs.cp(
    `${temporarySourceDirectory}/extensions/${extensionName}/dist/extension.js`,
    `${toDirectory}/extensions/${extensionName}`
  )
  shelljs.cp(
    `${temporarySourceDirectory}/extensions/${extensionName}/package.json`,
    `${toDirectory}/extensions/${extensionName}`
  )
}

signale.success('Done building Sourcegraph extensions')
