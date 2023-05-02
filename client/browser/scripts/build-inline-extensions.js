const path = require('path')

const { buildCodeIntelExtensions } = require('../../shared/dev/buildCodeIntelExtensions')

const pathToExtensionBundles = path.join(process.cwd(), 'build', 'extensions')

buildCodeIntelExtensions({ pathToExtensionBundles, revision: 'v5.0.1' })
