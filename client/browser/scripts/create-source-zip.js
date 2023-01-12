const path = require('path')

const shelljs = require('shelljs')
const signale = require('signale')

/**
 * Purpose of this script: create a source code zip that can be used shared and
 * used to produce a build of the browser extensions.
 *
 * This makes it possible to provide the complete source code to the maintainers
 * of extension registries, in the cases where providing the source code is a
 * pre-requisite to getting approval for publishing.
 *
 * This script fetches a fresh copy of the repository, identified by commit ID,
 * instead of zipping the current working directory. This is done instead for
 * these reasons:
 *
 * - To avoid zipping any changes in the current tree, or any existing build
 * artifacts.
 *
 * - To be able to exclude some unnecessary directories by deleting them before
 * zipping.
 *
 */

// Configuration

/**
 * Set the commitId to build from a given revision. Because it will be
 * downloaded from GitHub, this revision needs to exist there.
 */
let commitId = 'cedc530'

/**
 * If true, code intel extensions will be fetched and included in the zip, so
 * that they can be reviewed as part of the source code. This included copy will
 * be used when building. If not included, the code intel extensions will be
 * automatically downloaded during the build step.
 */
const includeCodeIntelExtensions = true

/**
 * Inside the zip, this will be the name of the root directory.
 */
const rootDirectoryNameForZip = 'sourcegraph-source'
const buildBashFile = 'build-ff.sh'

const revisionParseResult = shelljs.exec(`git rev-parse ${commitId}`, { silent: true })
if (revisionParseResult.code === 1) {
  signale.await('There was a problem using this commit id')
  process.exit(1)
}
commitId = revisionParseResult.stdout.trim()

// Clean up
shelljs.rm('-f', 'sourcegraph.zip')
shelljs.rm('-rf', rootDirectoryNameForZip)
shelljs.rm('-rf', `sourcegraph-${commitId}/`)

signale.await(`Downloading sourcegraph/sourcegraph at revision ${commitId}`)
shelljs.exec(
  `curl -Ls https://github.com/sourcegraph/sourcegraph/archive/${commitId}.zip -o sourcegraph-downloaded.zip`
)
shelljs.exec('unzip -q sourcegraph-downloaded.zip')
shelljs.rm('-f', 'sourcegraph-downloaded.zip')
shelljs.mv(`sourcegraph-${commitId}`, rootDirectoryNameForZip)
signale.success('Downloaded and unzipped sourcegraph/sourcegraph repository')

if (includeCodeIntelExtensions) {
  shelljs.pushd(rootDirectoryNameForZip)
  shelljs.exec('pnpm install')
  shelljs.exec('pnpm --filter @sourcegraph/browser run build-inline-extensions')
  shelljs.popd()
}

// Copy build script to main dir
shelljs.cp({}, path.join(process.cwd(), 'scripts', buildBashFile), path.join(rootDirectoryNameForZip, buildBashFile))

// Delete all unnecessary directories
shelljs.pushd(rootDirectoryNameForZip)
shelljs.rm('-rf', ['dev', 'doc', 'docker-images', 'enterprise', 'internal', 'migrations', 'monitoring', 'node_modules'])
shelljs.popd()

signale.await('Producing sourcegraph.zip')
shelljs.exec(`zip -qr sourcegraph.zip ${rootDirectoryNameForZip}`)
shelljs.rm('-rf', rootDirectoryNameForZip)
signale.success('Done producing sourcegraph.zip')
