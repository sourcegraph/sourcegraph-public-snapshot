import { cleanEnv, str } from 'envalid'

export const ENVIRONMENT_CONFIG = cleanEnv(process.env, {
    SRC_ACCESS_TOKEN: str(),
})

export const DEFAULTS = {
    codebase: 'github.com/sourcegraph/sourcegraph',
    serverEndpoint: 'https://sourcegraph.sourcegraph.com',
    contextType: 'blended',
    debug: 'development',
} as const
