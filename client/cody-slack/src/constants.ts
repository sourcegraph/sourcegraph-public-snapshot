import { cleanEnv, str, num } from 'envalid'
import { HNSWLib } from 'langchain/vectorstores/hnswlib'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'

export const ENVIRONMENT_CONFIG = cleanEnv(process.env, {
    PORT: num({ default: 3000 }),

    // OPENAI_API_KEY: str(),
    SOURCEGRAPH_ACCESS_TOKEN: str(),

    GITHUB_TOKEN: str(),

    SLACK_APP_TOKEN: str(),
    SLACK_BOT_TOKEN: str(),
    SLACK_SIGNING_SECRET: str(),
})

export const DEFAULT_APP_SETTINGS = {
    serverEndpoint: 'https://sourcegraph.sourcegraph.com',
    debug: 'development',
    contextType: 'blended',
} as const

export const DEFAULT_CODEBASES = [
    'github.com/sourcegraph/sourcegraph',
    'github.com/sourcegraph/handbook',
    'github.com/sourcegraph/about',
] as const

export type CodebaseContexts = Record<typeof DEFAULT_CODEBASES[number], CodebaseContext>
export interface AppContext {
    codebaseContexts: CodebaseContexts
    vectorStore: HNSWLib
}
