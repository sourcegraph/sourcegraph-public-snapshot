const path = require('path')

const { buildCodeIntelExtensions } = require('../../shared/src/extensions/buildCodeIntelExtensions')

const pathToExtensionBundles = path.join(process.cwd(), 'build', 'extensions')
const pathToRevisionFile = path.join(process.cwd(), 'sourcegraph-extension-bundles-revision.txt')
const pathToDistributionRevisionFile = path.join(pathToExtensionBundles, 'revision.txt')

buildCodeIntelExtensions({ pathToExtensionBundles, pathToRevisionFile, pathToDistributionRevisionFile })
