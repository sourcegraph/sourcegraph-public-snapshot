const { esbuildBuildOptions } = require('./esbuild.js')

module.exports = esbuildBuildOptions(process.env.NODE_ENV === 'development' ? 'dev' : 'prod')
