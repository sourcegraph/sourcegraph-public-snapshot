import { cleanEnv, str, num } from 'envalid'

export const ENVIRONMENT_CONFIG = cleanEnv(process.env, {
    PORT: num({ default: 3000 }),

    // OPENAI_API_KEY: str(),
    SOURCEGRAPH_ACCESS_TOKEN: str(),

    SLACK_APP_TOKEN: str(),
    SLACK_BOT_TOKEN: str(),
    SLACK_SIGNING_SECRET: str(),
})

export const DEFAULT_APP_SETTINGS = {
    codebase: 'github.com/sourcegraph/sourcegraph',
    serverEndpoint: 'https://sourcegraph.sourcegraph.com',
    contextType: 'blended',
    debug: 'development',
} as const
