import { SOLUTION_TOKEN_LENGTH } from '@sourcegraph/cody-shared/src/prompt/constants'

import { Message } from '../sourcegraph-api'
import { SourcegraphCompletionsClient, CompletionParameters } from '../sourcegraph-api/completions'
import { CompletionCallbacks } from '../sourcegraph-api/completions/types'

const DEFAULT_CHAT_COMPLETION_PARAMETERS: Omit<CompletionParameters, 'messages'> = {
    temperature: 0.2,
    maxTokensToSample: SOLUTION_TOKEN_LENGTH,
    topK: -1,
    topP: -1,
}

// Character length of this preamble is 806 chars or ~230 tokens (at a conservative rate of 3.5 chars per token).
// If this is modified, then `PROMPT_PREAMBLE_LENGTH` in prompt.ts should be updated.
const preamble: Message[] = [
    {
        speaker: 'human',
        text: `You are Cody, an AI-powered coding assistant created by Sourcegraph that performs the following actions:
- Answer general programming questions
- Answer questions about code that I have provided to you
- Generate code that matches a written description
- Explain what a section of code does

In your responses, you should obey the following rules:
- Be as brief and concise as possible without losing clarity
- Any code snippets should be markdown-formatted (placed in-between triple backticks like this "\`\`\`").
- Answer questions only if you know the answer or can make a well-informed guess. Otherwise, tell me you don't know, and tell me what context I need to provide to you in order for you to answer the question.
- Do not reference any file names or URLs, unless you are sure they exist.`,
    },
    {
        speaker: 'assistant',
        text: 'Understood. I am Cody, an AI-powered coding assistant created by Sourcegraph and will follow the rules above',
    },
]

export class ChatClient {
    constructor(private completions: SourcegraphCompletionsClient) {}

    public chat(messages: Message[], cb: CompletionCallbacks): () => void {
        const isLastMessageFromHuman = messages.length > 0 && messages[messages.length - 1].speaker === 'human'
        const augmentedMessages = isLastMessageFromHuman
            ? messages.concat([{ speaker: 'assistant', text: '' }])
            : messages

        return this.completions.stream(
            { messages: [...preamble, ...augmentedMessages], ...DEFAULT_CHAT_COMPLETION_PARAMETERS },
            cb
        )
    }
}
