/* eslint-disable no-sync */
const fs = require('fs')
const path = require('path')

const shelljs = require('shelljs')
const signale = require('signale')

const bundlesRepoName = 'sourcegraph-extensions-bundles'

/**
 * Clones https://github.com/sourcegraph/sourcegraph-extensions-bundles repo and copies programming languages
 * extensions bundles to the specified path. These bundles may be included into IDE/browser extensions bundles
 * in order not to require access to the extensions registry for the code intel features to work.
 */
function buildCodeIntelExtensions({ pathToExtensionBundles, pathToRevisionFile, pathToDistributionRevisionFile }) {
  const revisionFileContent = fs.existsSync(pathToRevisionFile) && fs.readFileSync(pathToRevisionFile).toString()
  const revision = revisionFileContent ? revisionFileContent.trim() : ''
  if (!revision) {
    // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
    signale.fatal(`Couldn't find "${bundlesRepoName}" revision to fetch at ${pathToRevisionFile}`)
    return
  }

  const currentRevision =
    fs.existsSync(pathToDistributionRevisionFile) && fs.readFileSync(pathToDistributionRevisionFile).toString()
  if (currentRevision === revision) {
    signale.success(`Code-intel-extensions for revision "${revision}" are already bundled.`)
    return
  }

  signale.info(`Did not find an existing code-intel-extensions bundles matching revision ${revision}.`)

  signale.watch('Fetching code intel extensions bundles...')
  shelljs.exec(`curl -OLs https://github.com/sourcegraph/${bundlesRepoName}/archive/${revision}.zip`)

  // when repo archive is unpacked the leading 'v' from tag is trimmed: v1.0.0.zip => sourcegraph-extensions-bundles-1.0.0
  const bundlesRepoDirectoryName = `${bundlesRepoName}-${revision.replace(/^v/g, '')}`

  // Remove existing repo and bundles directory in case of an interrupted process.
  shelljs.rm('-rf', bundlesRepoDirectoryName)
  shelljs.rm('-rf', pathToExtensionBundles)

  // Unzip bundles repo archive, which creates a new directory: sourcegraph-extensions-bundles-{revision}
  shelljs.exec(`unzip -q ${revision}.zip`)
  // Remove bundles repo archive
  shelljs.exec(`rm ${revision}.zip`)

  const codeIntelExtensionIds = [] // list of cloned code intel extension ids, e.g. [..., 'sourcegraph/typescript', ...]
  const content = fs.readdirSync(path.join(bundlesRepoDirectoryName, 'bundles'), { withFileTypes: true })

  for (const item of content) {
    if (!item.isDirectory()) {
      continue
    }

    const extensionName = item.name // kebab-case extension name, e.g. 'sourcegraph-typescript'

    const bundlePath = path.join(bundlesRepoDirectoryName, 'bundles', extensionName)
    const files = fs.readdirSync(bundlePath)

    if (['package.json', `${extensionName}.js`].some(file => !files.includes(file))) {
      // does not look like a valid bundle directory
      continue
    }

    let isProgrammingLanguageExtension = false
    try {
      const packageJsonContent = JSON.parse(fs.readFileSync(path.join(bundlePath, 'package.json')).toString())
      isProgrammingLanguageExtension = packageJsonContent.categories?.includes('Programming languages')
    } catch {
      // couldn't parse package.json
      continue
    }

    if (!isProgrammingLanguageExtension) {
      continue
    }

    shelljs.mkdir('-p', path.join(pathToExtensionBundles, extensionName))
    shelljs.exec(
      `cp ${path.join(bundlePath, 'package.json')} ${path.join(pathToExtensionBundles, extensionName, 'package.json')}`
    )
    shelljs.exec(
      `cp ${path.join(bundlePath, `${extensionName}.js`)} ${path.join(
        pathToExtensionBundles,
        extensionName,
        'extension.js'
      )}`
    )

    codeIntelExtensionIds.push(extensionName.replace(/^sourcegraph-/, 'sourcegraph/'))
  }

  // Remove bundles repo directory and archive
  shelljs.exec(`rm -rf ${bundlesRepoDirectoryName}`)

  // Save sourcegraph-extensions-bundles revision not to refetch it on the next calls if the revision doesn't change
  fs.writeFileSync(pathToDistributionRevisionFile, revision)

  // Save extension IDs of the copied bundles
  fs.writeFileSync(path.join(process.cwd(), 'code-intel-extensions.json'), JSON.stringify(codeIntelExtensionIds))

  signale.success('Code intel extensions bundles successfully copied.')
}

module.exports = { buildCodeIntelExtensions }
