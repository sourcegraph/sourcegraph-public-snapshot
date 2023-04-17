import { IncomingMessage } from 'http'

import { Configuration, OpenAIApi } from 'openai'

import { SourcegraphCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/client'
import {
    CodeCompletionResponse,
    CompletionCallbacks,
    CompletionParameters,
    Message,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

export class OpenAICompletionsClient extends SourcegraphCompletionsClient {
    private openai: OpenAIApi

    constructor(
        protected apiKey: string,
        instanceUrl: string = '',
        protected accessToken: string = '',
        protected mode: 'development' | 'production' = 'production'
    ) {
        super(instanceUrl, accessToken, mode)

        const configuration = new Configuration({
            apiKey: this.apiKey,
        })
        this.openai = new OpenAIApi(configuration)
    }

    public stream(params: CompletionParameters, cb: CompletionCallbacks) {
        this.openai
            .createChatCompletion(
                {
                    model: 'gpt-3.5-turbo',
                    messages: params.messages
                        .filter(
                            (message): message is Omit<Message, 'text'> & Required<Pick<Message, 'text'>> =>
                                message.text !== undefined
                        )
                        .map(message => {
                            return {
                                role: message.speaker === 'human' ? 'user' : 'assistant',
                                content: message.text,
                            }
                        }),
                    stream: true,
                },
                { responseType: 'stream' }
            )
            .then(response => {
                const stream = response.data as unknown as IncomingMessage

                let modelResponseText = ''

                stream.on('data', (chunk: Buffer) => {
                    // Split messames in the event stream.
                    const payloads = chunk.toString().split('\n\n')

                    for (const payload of payloads) {
                        if (payload.includes('[DONE]')) {
                            return
                        }

                        if (payload.startsWith('data:')) {
                            const data = payload.replaceAll(/(\n)?^data:\s*/g, '') // in case there's multiline data event

                            try {
                                const delta = JSON.parse(data.trim())
                                const newTextChunk = delta.choices[0].delta?.content

                                if (newTextChunk) {
                                    modelResponseText += newTextChunk
                                    cb.onChange(modelResponseText)
                                }
                            } catch (error) {
                                console.log(
                                    `Error with JSON.parse: ${chunk.toString()}\nPayload: ${payload};\nError: ${error}`
                                )
                                cb.onError(error)
                            }
                        }
                    }
                })

                stream.on('error', e => cb.onError(e.message))
                stream.on('end', () => cb.onComplete())
            })
            .catch(console.error)

        return () => {}
    }

    public complete(): Promise<CodeCompletionResponse> {
        throw new Error('SourcegraphBrowserCompletionsClient.complete not implemented')
    }
}
