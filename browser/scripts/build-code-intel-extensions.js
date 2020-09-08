const shelljs = require('shelljs')
const path = require('path')
// eslint-disable-next-line id-length
const os = require('os')
const signale = require('signale')

/**
 * Name of the extension to build, from code-intel-extensions. This script
 * currently builds only a single extension.
 */
const extensionName = 'template'

/**
 * Revision of the github.com/sourcegraph/code-intel-extensions repo to build
 * from.
 *
 * We freeze it at a particular revision because the XPI needs to be
 * reproducible. The particular revision was chosen as the latest commit on
 * master at the time of writing and should be updated accordingly.
 */
const codeIntelExtensionsRepoRevision = 'ff747b5bcdbbd5809ddd2737480d2ced88d179d5'

const toDirectory = path.join(process.cwd(), 'build')

signale.await('Building Sourcegraph extensions for Firefox add-on')

const temporaryCloneDirectory = path.join(os.tmpdir(), 'code-intel-extensions')
shelljs.mkdir('-p', temporaryCloneDirectory)

// Get code-intel-extensions source snapshot and build
shelljs.pushd(temporaryCloneDirectory)
shelljs.exec(
  `curl -OLs https://github.com/sourcegraph/code-intel-extensions/archive/${codeIntelExtensionsRepoRevision}.zip`
)

// Prepare code-intel-extensions for build:
// - Clean: remove old directories
shelljs.rm('-rf', 'code-intel-extensions')
shelljs.rm('-rf', `code-intel-extensions-${codeIntelExtensionsRepoRevision}`)

// - Unzip it, which creates a new directory: code-intel-extensions-{rev}
shelljs.exec(`unzip -q ${codeIntelExtensionsRepoRevision}.zip`)

// - Rename directory to remove revision suffix
shelljs.mv(`code-intel-extensions-${codeIntelExtensionsRepoRevision}`, 'code-intel-extensions')

// Install dependencies and build the specified code intel extension
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

signale.success('Done building Sourcegraph extensions')
