const path = require('path')

const { buildCodeIntelExtensions } = require('../../shared/dev/buildCodeIntelExtensions')

const pathToExtensionBundles = path.join(process.cwd(), 'dist', 'extensions')

buildCodeIntelExtensions({ pathToExtensionBundles, revision: 'v3.41.1' })
