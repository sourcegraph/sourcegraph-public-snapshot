import { SOLUTION_TOKEN_LENGTH } from '@sourcegraph/cody-shared/src/prompt/constants'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import {
    CompletionParameters,
    CompletionCallbacks,
    Message,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { DEFAULT_APP_SETTINGS, ENVIRONMENT_CONFIG } from '../constants'

import { OpenAICompletionsClient } from './openai-completions-client'

const { SOURCEGRAPH_ACCESS_TOKEN } = ENVIRONMENT_CONFIG

const DEFAULT_CHAT_COMPLETION_PARAMETERS: Omit<CompletionParameters, 'messages'> = {
    temperature: 0.2,
    maxTokensToSample: SOLUTION_TOKEN_LENGTH,
    topK: -1,
    topP: -1,
}

const completionsClient = getCompletionsClient()

export function streamCompletions(messages: Message[], cb: CompletionCallbacks) {
    return completionsClient.stream({ messages, ...DEFAULT_CHAT_COMPLETION_PARAMETERS }, cb)
}

function getCompletionsClient() {
    const { OPENAI_API_KEY } = process.env

    if (OPENAI_API_KEY) {
        return new OpenAICompletionsClient(OPENAI_API_KEY)
    }

    return new SourcegraphNodeCompletionsClient(
        DEFAULT_APP_SETTINGS.serverEndpoint,
        SOURCEGRAPH_ACCESS_TOKEN,
        DEFAULT_APP_SETTINGS.debug
    )
}
