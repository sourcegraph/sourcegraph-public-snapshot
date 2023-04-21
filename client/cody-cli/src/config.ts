import { cleanEnv, str } from 'envalid'

export const ENVIRONMENT_CONFIG = cleanEnv(
    process.env,
    {
        SRC_ACCESS_TOKEN: str(),
    },
    {
        reporter: ({ errors, env }) => {
            if (errors.SRC_ACCESS_TOKEN) {
                console.error('SRC_ACCESS_TOKEN not set')
            }
        },
    }
)

export const DEFAULTS = {
    codebase: 'github.com/sourcegraph/sourcegraph',
    serverEndpoint: 'https://sourcegraph.sourcegraph.com',
    contextType: 'blended',
    debug: 'development',
} as const
