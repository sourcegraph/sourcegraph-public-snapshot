import { cleanEnv, str } from 'envalid'

export const ENVIRONMENT_CONFIG = cleanEnv(process.env, {
    SOURCEGRAPH_ACCESS_TOKEN: str(),
    SOURCEGRAPH_SERVER_ENDPOINT: str(),
})

export const DEFAULT_APP_SETTINGS = {
    contextType: 'blended',
    debug: 'development',
    codebase: 'github.com/sourcegraph/sourcegraph', // TODO: GITHUB_REPOSITORY
} as const
