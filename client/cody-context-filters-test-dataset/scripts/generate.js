const path = require('path')
const shelljs = require('shelljs')
const semver = require('semver')
const signale = require('signale')

;(() => {
  const datasetPath = path.join(
    __dirname,
    '../../../cmd/frontend/internal/codycontext/testdata/enterprise_filter_test_data.json'
  )
  const destinationPath = path.join(__dirname, '../dataset.json')

  const copyDatasetFileResult = shelljs.cp(datasetPath, destinationPath)
  if (copyDatasetFileResult.code !== 0) {
    signale.error('Failed to copy test dataset:', copyDatasetFileResult.stderr)
    shelljs.exit(1)
  }

  const readDatasetContent = shelljs.cat(destinationPath)
  if (readDatasetContent.code !== 0) {
    signale.error('Failed to read dataset content:', readDatasetContent.stderr)
    shelljs.exit(1)
  }

  try {
    JSON.parse(readDatasetContent.stdout)
  } catch (e) {
    signale.error('Failed to parse dataset content as JSON:', e)
    shelljs.exit(1)
  }

  const readPackageJSONContent = shelljs.cat(path.join(__dirname, '../package.json'))
  let versionFromPackageJSON
  try {
    versionFromPackageJSON = JSON.parse(readPackageJSONContent.stdout).version
  } catch (e) {
    signale.error('Failed to parse package.json:', e)
    shelljs.exit(1)
  }

  const version = semver.valid(versionFromPackageJSON)
  if (!version) {
    signale.error(
      `Invalid version in package.json: ${JSON.stringify(
        versionFromPackageJSON
      )}. Versions must be valid semantic version strings.`
    )
    shelljs.exit(1)
  }
})()
