const shelljs = require('shelljs')
const path = require('path')
const signale = require('signale')
const bundledCodeIntelExtensionsConfig = require('../bundled-code-intel-extensions.json')
const temporarySourceDirectory = path.join(process.cwd(), 'code-intel-extensions')

/**
 * Revision of the github.com/sourcegraph/code-intel-extensions repo to build
 * from.
 *
 * We freeze it at a particular revision because the XPI needs to be
 * reproducible. The revision should be updated whenever a new version of the
 * browser extension is being built for release.
 */
const codeIntelExtensionsRepoRevision = bundledCodeIntelExtensionsConfig.revision

signale.await('Fetching code-intel-extensions')

// Refetch
shelljs.rm('-rf', temporarySourceDirectory)

shelljs.mkdir('-p', temporarySourceDirectory)

// Get code-intel-extensions source snapshot and build
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
