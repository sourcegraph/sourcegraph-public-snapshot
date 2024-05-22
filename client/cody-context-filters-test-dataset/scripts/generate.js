const path = require('path')
const shelljs = require('shelljs')
const semver = require('semver')
const signale = require('signale')

;(() => {
  signale.await('Copying test dataset...')
  const datasetSource = path.join(
    __dirname,
    '../../../cmd/frontend/internal/codycontext/testdata/enterprise_filter_test_data.json'
  )
  const datasetDest = path.join(__dirname, '../dataset.json')
  const packageJSONPath = path.join(__dirname, '../package.json')

  const copyDatasetFileResult = shelljs.cp(datasetSource, datasetDest)
  if (copyDatasetFileResult.code !== 0) {
    signale.fatal('Failed to copy test dataset:', copyDatasetFileResult.stderr)
    shelljs.exit(1)
  }

  const readDatasetContent = shelljs.cat(datasetDest)
  if (readDatasetContent.code !== 0) {
    signale.fatal('Failed to read dataset content:', readDatasetContent.stderr)
    shelljs.exit(1)
  }

  try {
    JSON.parse(readDatasetContent.stdout)
  } catch (e) {
    signale.fatal('Failed to parse dataset content as JSON:', e)
    shelljs.exit(1)
  }

  const readPackageJSONContent = shelljs.cat(packageJSONPath)
  let versionFromPackageJSON
  try {
    versionFromPackageJSON = JSON.parse(readPackageJSONContent.stdout).version
  } catch (e) {
    signale.fatal('Failed to parse package.json:', e)
    shelljs.exit(1)
  }

  const version = semver.valid(versionFromPackageJSON)
  if (!version) {
    signale.fatal(
      `Invalid version in package.json: ${JSON.stringify(
        versionFromPackageJSON
      )}. Versions must be valid semantic version strings.`
    )
    shelljs.exit(1)
  }

  signale.success('Test dataset created successfully!')
})()
