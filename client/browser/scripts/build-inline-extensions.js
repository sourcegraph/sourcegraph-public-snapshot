const fs = require('fs')
const path = require('path')

const shelljs = require('shelljs')
const signale = require('signale')

const config = require('../bundled-code-intel-extensions.json')
const extensionNames = config.extensions || []

const toDirectory = path.join(process.cwd(), 'build')
const temporarySourceDirectory = path.join(process.cwd(), 'code-intel-extensions')
const pathToRevisionFile = path.join(process.cwd(), 'code-intel-extensions', 'revision.txt')

// Check if code-intel-extensions has already been fetched
const codeIntelExtensionsDirectoryExists = shelljs.test('-d', temporarySourceDirectory)
if (codeIntelExtensionsDirectoryExists && fs.existsSync(pathToRevisionFile) && fs.readFileSync(pathToRevisionFile).toString() === config.revision) {
  console.log('Found existing code-intel-extensions.')
} else {
  console.log('Did not find an existing code-intel-extensions. Running fetch-code-intel-extensions')
  console.log(fs.existsSync(pathToRevisionFile) ? fs.readFileSync(pathToRevisionFile).toString() : 'not found', config.revision)
  shelljs.exec('yarn run fetch-code-intel-extensions')
  // Save revision info
  shelljs.exec(`touch ${pathToRevisionFile}`)
  shelljs.exec(`echo ${config.revision} >> ${pathToRevisionFile}`)
}

// Install dependencies
shelljs.exec('yarn --cwd code-intel-extensions install')

// Build individual extensions
for (const extensionName of extensionNames) {
  const extensionDirectory = `${temporarySourceDirectory}/extensions/${extensionName}`
  if (!shelljs.test('-d', extensionDirectory)) {
    console.error(`Code intel extension "${extensionName}" was not found for bundling.`)
    console.error(`Expected extension directory: ${extensionDirectory}`)
    process.exit(1)
  }
  shelljs.exec(`yarn --cwd ${temporarySourceDirectory}/extensions/${extensionName} run build`)
}

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
