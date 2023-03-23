const { createDevelopmentServerConfig } = require('./dev/server/bazel.server')
const base = require('./webpack.bazel.config')

module.exports = {
  ...base,
  devServer: createDevelopmentServerConfig(),
}
