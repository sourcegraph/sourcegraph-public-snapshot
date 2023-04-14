// Simple express server to serve a static production build of the prototype
import express from 'express'
import { createProxyMiddleware } from 'http-proxy-middleware'

const app = express()
const port = 4173

// Serve prototype files
app.use(express.static('./build', { fallthrough: true }))
// Proxy API, stream and other specific endpoints to Sourcegraph instance
app.use(
  /^\/(sign-in|.assets|-|.api|search\/stream)/,
  createProxyMiddleware({ target: process.env['SOURCEGRAPH_API_HOST'], changeOrigin: true, secure: false })
)
// Fallback route to make SPA work for any URL on cold load
app.all('*', (_req, res) => res.sendFile('index.html', { root: './build' }))

app.listen(port, () => {
  console.log(`Listening on port ${port}`)
})
