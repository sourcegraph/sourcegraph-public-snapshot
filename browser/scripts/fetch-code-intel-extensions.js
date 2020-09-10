const shelljs = require('shelljs')
const path = require('path')
// eslint-disable-next-line id-length
const os = require('os')
const signale = require('signale')
const bundledCodeIntelExtensionsConfig = require('../bundled-code-intel-extensions.json')
const temporarySourceDirectory = path.join('./', 'code-intel-extensions')
const toDirectory = path.join(process.cwd(), 'build')

/**
 * Revision of the github.com/sourcegraph/code-intel-extensions repo to build
 * from.
 *
 * We freeze it at a particular revision because the XPI needs to be
 * reproducible. The particular revision was chosen as the latest commit on
 * master at the time of writing and should be updated accordingly.
 */
const codeIntelExtensionsRepoRevision = bundledCodeIntelExtensionsConfig.revision
const extensionNames = bundledCodeIntelExtensionsConfig.extensions

signale.await('Building Sourcegraph extensions for Firefox add-on')

shelljs.mkdir('-p', temporarySourceDirectory)

// Get code-intel-extensions source snapshot and build
// shelljs.pushd(temporarySourceDirectory)
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

shelljs.rm('-f', `"${codeIntelExtensionsRepoRevision}.zip"`)
