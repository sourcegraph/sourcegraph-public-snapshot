import { SOLUTION_TOKEN_LENGTH } from '../prompt/constants'
import { Message } from '../sourcegraph-api'
import { SourcegraphCompletionsClient, CompletionParameters } from '../sourcegraph-api/completions'
import { CompletionCallbacks } from '../sourcegraph-api/completions/types'

const DEFAULT_CHAT_COMPLETION_PARAMETERS: Omit<CompletionParameters, 'messages'> = {
    temperature: 0.2,
    maxTokensToSample: SOLUTION_TOKEN_LENGTH,
    topK: -1,
    topP: -1,
}

export class ChatClient {
    constructor(private completions: SourcegraphCompletionsClient) {}

    public chat(messages: Message[], cb: CompletionCallbacks): () => void {
        const isLastMessageFromHuman = messages.length > 0 && messages[messages.length - 1].speaker === 'human'
        const augmentedMessages = isLastMessageFromHuman
            ? messages.concat([{ speaker: 'assistant', text: '' }])
            : messages

        return this.completions.stream({ messages: augmentedMessages, ...DEFAULT_CHAT_COMPLETION_PARAMETERS }, cb)
    }
}
