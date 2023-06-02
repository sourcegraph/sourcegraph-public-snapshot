import { IncomingMessage } from 'http'

import { Configuration, OpenAIApi } from 'openai'

import { SourcegraphCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/client'
import {
    CompletionCallbacks,
    CompletionParameters,
    Message,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

export class OpenAICompletionsClient implements Pick<SourcegraphCompletionsClient, 'stream'> {
    private openai: OpenAIApi

    constructor(protected apiKey: string) {
        const configuration = new Configuration({
            apiKey: this.apiKey,
        })

        this.openai = new OpenAIApi(configuration)
    }

    public stream(params: CompletionParameters, cb: CompletionCallbacks) {
        this.openai
            .createChatCompletion(
                {
                    // TODO: manage prompt length
                    model: 'gpt-4',
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
                let buffer = ''

                stream.on('data', (chunk: Buffer) => {
                    // Split messages in the event stream.
                    buffer += chunk.toString()
                    const payloads = buffer.split('\n\n')

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

                                buffer = buffer.slice(Math.max(0, buffer.indexOf(payload) + payload.length))
                            } catch (error) {
                                if (error instanceof SyntaxError && buffer.length > 0) {
                                    // Incomplete JSON string, wait for more data
                                    continue
                                } else {
                                    console.log(
                                        `Error with JSON.parse: ${chunk.toString()}\nPayload: ${payload};\nError: ${error}`
                                    )
                                    cb.onError(error)
                                }
                            }
                        }
                    }
                })

                stream.on('error', e => {
                    console.error('OpenAI stream failed', e)
                    cb.onError(e.message)
                })
                stream.on('end', () => cb.onComplete())
            })
            .catch(console.error)

        return () => {}
    }
}
