import { ANSWER_TOKENS } from '@sourcegraph/cody-shared/src/prompt/constants'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import {
    CompletionParameters,
    CompletionCallbacks,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

const DEFAULT_CHAT_COMPLETION_PARAMETERS: Omit<CompletionParameters, 'messages'> = {
    temperature: 0.2,
    maxTokensToSample: ANSWER_TOKENS,
    topK: -1,
    topP: -1,
}

export function streamCompletions(
    client: SourcegraphNodeCompletionsClient,
    messages: Message[],
    cb: CompletionCallbacks
) {
    return client.stream({ messages, ...DEFAULT_CHAT_COMPLETION_PARAMETERS }, cb)
}
