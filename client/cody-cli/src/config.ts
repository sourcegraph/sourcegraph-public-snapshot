import { cleanEnv, str } from 'envalid'

export function getEnvConfig(): { SRC_ACCESS_TOKEN: string } {
    return cleanEnv(process.env, {
        SRC_ACCESS_TOKEN: str(),
    })
}

export const DEFAULTS = {
    codebase: 'github.com/sourcegraph/sourcegraph',
    serverEndpoint: 'https://sourcegraph.sourcegraph.com',
    contextType: 'blended',
    debug: 'development',
} as const
