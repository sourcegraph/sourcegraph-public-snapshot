const path = require('path')

require('ts-node').register({
  transpileOnly: true,
  // Use config with "module": "commonjs" because not all modules involved in tasks are esnext modules.
  project: path.resolve(__dirname, './tsconfig.json'),
})

const { createProxyMiddleware } = require('http-proxy-middleware')

const { esbuildDevelopmentServer } = require('@sourcegraph/web/dev/esbuild/server')
const { BUILD_OPTIONS } = require('./esbuild/build')
const {
  DEV_SERVER_LISTEN_ADDR,
  DEV_SERVER_PROXY_TARGET_ADDR,
} = require('@sourcegraph/web/dev/utils')
const {serve} = require('esbuild')


const developmentServer = () => serve(
  { host: DEV_SERVER_LISTEN_ADDR.host, port:8077, servedir: 'static' },
  BUILD_OPTIONS
)


module.exports = {
  developmentServer,
}
