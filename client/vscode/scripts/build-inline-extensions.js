const fs = require('fs')
const path = require('path')

const shelljs = require('shelljs')
const signale = require('signale')

const { extensions: extensionNames = [], revision } = require('../bundled-code-intel-extensions.json')

const toDirectory = path.join(process.cwd(), 'dist')
const temporarySourceDirectory = path.join(process.cwd(), 'code-intel-extensions')
const pathToRevisionFile = path.join(process.cwd(), 'code-intel-extensions', 'revision.txt')
const pathToDistRevisionFile = path.join(toDirectory, 'extensions', 'revision.txt')

// Check if code-intel-extensions has already been fetched
if (fs.existsSync(pathToRevisionFile) && fs.readFileSync(pathToRevisionFile).toString() === revision) {
  console.log('Found existing code-intel-extensions.')
} else {
  console.log('Did not find an existing code-intel-extensions. Running fetch-code-intel-extensions')
  shelljs.exec('yarn run fetch-code-intel-extensions')
  fs.writeFileSync(pathToRevisionFile, revision)
}

// Check if dependencies have already been built
if (fs.existsSync(pathToDistRevisionFile) && fs.readFileSync(pathToDistRevisionFile).toString() === revision) {
  console.log('Extensions are already buillt.')
} else {
  // Install dependencies
  shelljs.exec('yarn --cwd code-intel-extensions install')

  // Generate individual extensions
  shelljs.exec('yarn --cwd code-intel-extensions generate')

  console.log('Generating code-intel extensions bundles...')

  for (const extensionName of extensionNames) {
    const extensionDir = path.join(temporarySourceDirectory, `generated-${extensionName}`)

    const pathToPackageJSON = path.join(extensionDir, 'package.json')

    shelljs.exec(`yarn --cwd ${extensionDir} build`)

    shelljs.mkdir('-p', `${toDirectory}/extensions/${extensionName}`)
    shelljs.exec(`cp ${pathToPackageJSON} ${toDirectory}/extensions/${extensionName}/package.json`)
    shelljs.exec(
      `cp ${path.join(extensionDir, 'dist', 'extension.js')} ${path.join(
        toDirectory,
        'extensions',
        extensionName,
        `extension.js`
      )}`
    )
    console.log(`Successfully generated sourcegraph/${extensionName} bundle.`)
  }

  fs.writeFileSync(pathToDistRevisionFile, revision)

  signale.success('Done building Sourcegraph extensions')
}
