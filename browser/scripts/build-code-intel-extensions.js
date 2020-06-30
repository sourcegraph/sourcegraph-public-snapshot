const shelljs = require('shelljs')
const path = require('path')
// eslint-disable-next-line id-length
const os = require('os')
const signale = require('signale')

const extensionName = 'template'

const toDirectory = path.join(process.cwd(), 'build')

signale.await('Building extensions for Firefox addon')

const temporaryCloneDirectory = path.join(os.tmpdir(), 'code-intel-extensions')
shelljs.mkdir('-p', temporaryCloneDirectory)

// Clone and build
shelljs.pushd(temporaryCloneDirectory)
shelljs.exec('git clone git@github.com:sourcegraph/code-intel-extensions || (cd code-intel-extensions; git pull)')
shelljs.exec('yarn --cwd code-intel-extensions install')
shelljs.exec(`yarn --cwd code-intel-extensions/extensions/${extensionName} run build`)
shelljs.popd()

// Copy extension manifest (package.json) and bundle (extension.js)
shelljs.mkdir('-p', `${toDirectory}/extensions/${extensionName}`)

shelljs.cp(
  `${temporaryCloneDirectory}/code-intel-extensions/extensions/${extensionName}/dist/extension.js`,
  `${toDirectory}/extensions/${extensionName}`
)
shelljs.cp(
  `${temporaryCloneDirectory}/code-intel-extensions/extensions/${extensionName}/package.json`,
  `${toDirectory}/extensions/${extensionName}`
)

signale.success('Done building extensions')
