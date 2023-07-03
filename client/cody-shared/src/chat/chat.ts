import { SOLUTION_TOKEN_LENGTH } from '../prompt/constants'
import { Message } from '../sourcegraph-api'
import type { SourcegraphCompletionsClient } from '../sourcegraph-api/completions/client'
import type { CompletionParameters, CompletionCallbacks } from '../sourcegraph-api/completions/types'

type ChatParameters = Omit<CompletionParameters, 'messages'>

const DEFAULT_CHAT_COMPLETION_PARAMETERS: ChatParameters = {
    temperature: 0.2,
    maxTokensToSample: SOLUTION_TOKEN_LENGTH,
    topK: -1,
    topP: -1,
}

const createTypewriter = (baseDelay: number, emit: (text: string) => void) => {
    let processedText = '';
    let interval: ReturnType<typeof setInterval> | undefined

    return (updatedText: string) => {
        const remainingChars = processedText.length - updatedText.length
        const dynamicDelay = Math.max(baseDelay / remainingChars, 5)

        if (interval) {
            clearInterval(interval)
            interval = undefined
        }

        interval = setInterval(() => {
            const nextChar = updatedText[processedText.length];
            processedText += nextChar;

            if (processedText.length === updatedText.length) {
                clearInterval(interval);
                interval = undefined
            }

            return emit(processedText);
        }, dynamicDelay)
    }
}

export class ChatClient {
    constructor(private completions: SourcegraphCompletionsClient) {}

    public chat(messages: Message[], cb: CompletionCallbacks, params?: Partial<ChatParameters>): () => void {
        const isLastMessageFromHuman = messages.length > 0 && messages[messages.length - 1].speaker === 'human'
        const augmentedMessages = isLastMessageFromHuman ? messages.concat([{ speaker: 'assistant' }]) : messages

        const writeChars = createTypewriter(100, cb.onChange)
        return this.completions.stream(
                {
                    ...DEFAULT_CHAT_COMPLETION_PARAMETERS,
                    ...params,
                    messages: augmentedMessages,
                },
                {
                    ...cb,
                    onChange: writeChars,
                    onComplete: () => {
                        console.log('OMG WE DONE----')
                        cb.onComplete()
                    }
                }
            )
    }
}
