const path = require('path')

const { fetchAndBuildCodeIntelExtensions } = require('../../shared/dev/buildCodeIntelExtensions')

const pathToExtensionBundles = path.join(process.cwd(), 'build', 'extensions')

fetchAndBuildCodeIntelExtensions({ pathToExtensionBundles, revision: 'v5.0.1' })
