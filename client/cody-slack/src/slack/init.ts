import { App } from '@slack/bolt'
import { WebClient, LogLevel } from '@slack/web-api'

import { ENVIRONMENT_CONFIG } from '../constants'

// Initialize the Bolt.js app and the WebClient using environment variables.
export const app = new App({
    appToken: ENVIRONMENT_CONFIG.SLACK_APP_TOKEN,
    token: ENVIRONMENT_CONFIG.SLACK_BOT_TOKEN,
    signingSecret: ENVIRONMENT_CONFIG.SLACK_SIGNING_SECRET,
    logLevel: LogLevel.INFO,
    customRoutes: [
        {
            path: '/healthz',
            method: ['GET'],
            handler: (req, res) => {
                res.writeHead(200)
                res.end(`Things are going just fine at ${req.headers.host}!`)
            },
        },
    ],
})

export const webClient = new WebClient(ENVIRONMENT_CONFIG.SLACK_BOT_TOKEN)
