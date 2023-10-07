import { createProxyMiddleware } from 'http-proxy-middleware'
import signale from 'signale'

import { getManifest } from '../esbuild/manifestPlugin'
import { esbuildDevelopmentServer } from '../esbuild/server'
import { ENVIRONMENT_CONFIG, getAPIProxySettings, getIndexHTML } from '../utils'

const { SOURCEGRAPH_API_URL, SOURCEGRAPH_HTTP_PORT } = ENVIRONMENT_CONFIG

interface DevelopmentServerInit {
    apiURL: string
}

async function startDevelopmentServer(): Promise<void> {
    signale.start('Starting dev server.', ENVIRONMENT_CONFIG)

    if (!SOURCEGRAPH_API_URL) {
        throw new Error('development.server.ts only supports *web-standalone* usage')
    }

    await startEsbuildDevelopmentServer({
        apiURL: SOURCEGRAPH_API_URL,
    })
}

async function startEsbuildDevelopmentServer({ apiURL }: DevelopmentServerInit): Promise<void> {
    const manifestFile = getManifest()
    const htmlPage = getIndexHTML({ manifestFile })

    const { proxyRoutes, ...proxyMiddlewareOptions } = getAPIProxySettings({
        apiURL,
        getLocalIndexHTML(jsContextScript) {
            return getIndexHTML({ manifestFile, jsContextScript })
        },
    })

    await esbuildDevelopmentServer({ host: '0.0.0.0', port: SOURCEGRAPH_HTTP_PORT }, app => {
        app.use(createProxyMiddleware(proxyRoutes, proxyMiddlewareOptions))
        app.get(/.*/, (_request, response) => {
            response.send(htmlPage)
        })
    })
}

startDevelopmentServer().catch(error => signale.error(error))
