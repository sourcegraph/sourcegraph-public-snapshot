const path = require('path')

require('ts-node').register({
  transpileOnly: true,
  // Use config with "module": "commonjs" because not all modules involved in tasks are esnext modules.
  project: path.resolve(__dirname, './tsconfig.json'),
})

const buildScripts = require('./scripts/build')

async function esbuild() {
  await buildScripts.build()
}

module.exports = { esbuild }
