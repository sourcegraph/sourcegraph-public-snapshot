const path = require('path')

require('ts-node').register({
  transpileOnly: true,
  // Use config within ./scripts/ to run scripts with ts-node
  project: path.resolve(__dirname, './scripts/tsconfig.json'),
})

const buildScripts = require('./scripts/build')

async function esbuild() {
  await buildScripts.build()
}

module.exports = { esbuild }
