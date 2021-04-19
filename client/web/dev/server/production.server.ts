import 'dotenv/config'

import historyApiFallback from 'connect-history-api-fallback'
import express from 'express'
import { createProxyMiddleware } from 'http-proxy-middleware'
import open from 'open'

import {
    PROXY_ROUTES,
    getAPIProxySettings,
    getCSRFTokenCookieMiddleware,
    environmentConfig,
    getCSRFTokenAndCookie,
    STATIC_ASSETS_PATH,
} from '../utils'

const { SOURCEGRAPH_API_URL, SOURCEGRAPH_HTTPS_PORT, SOURCEGRAPH_HTTPS_DOMAIN } = environmentConfig
const PROD_SERVER_URL = `http://${SOURCEGRAPH_HTTPS_DOMAIN}:${SOURCEGRAPH_HTTPS_PORT}`

async function startProductionServer(): Promise<void> {
    const { csrfContextValue, csrfCookieValue } = await getCSRFTokenAndCookie(SOURCEGRAPH_API_URL)
    console.log('Starting production server...', { ...environmentConfig, csrfContextValue, csrfCookieValue })

    const app = express()

    app.use(historyApiFallback())
    app.use(getCSRFTokenCookieMiddleware(csrfCookieValue))

    app.use(express.static(STATIC_ASSETS_PATH))
    app.use('/.assets', express.static(STATIC_ASSETS_PATH))

    app.use(
        PROXY_ROUTES,
        createProxyMiddleware(
            getAPIProxySettings({
                csrfContextValue,
                apiURL: SOURCEGRAPH_API_URL,
            })
        )
    )

    app.listen(SOURCEGRAPH_HTTPS_PORT, () => {
        console.log(`[PROD] Server: ${PROD_SERVER_URL}`)

        return open(`${PROD_SERVER_URL}/search`)
    })
}

startProductionServer().catch(error => console.error('Something went wrong :(', error))
