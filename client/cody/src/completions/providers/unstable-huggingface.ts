import { Completion } from '..'
import { logger } from '../../log'
import { isAbortError } from '../utils'

import { Provider, ProviderConfig, ProviderOptions } from './provider'

interface UnstableHuggingFaceOptions {
    serverEndpoint: string
    accessToken: null | string
}

const PROVIDER_IDENTIFIER = 'huggingface'
const STOP_WORD = '<|endoftext|>'

export class UnstableHuggingFaceProvider extends Provider {
    private serverEndpoint: string
    private accessToken: null | string

    constructor(options: ProviderOptions, unstableHuggingFaceOptions: UnstableHuggingFaceOptions) {
        super(options)
        this.serverEndpoint = unstableHuggingFaceOptions.serverEndpoint
        this.accessToken = unstableHuggingFaceOptions.accessToken
    }

    public async generateCompletions(abortSignal: AbortSignal): Promise<Completion[]> {
        // TODO: Add context and language
        const prompt = `<fim_prefix>${this.prefix}<fim_suffix>${this.suffix}<fim_middle>`

        const request = {
            inputs: prompt,
            parameters: {
                num_return_sequences: 1,
                max_new_tokens: this.multilineMode === null ? 40 : 128,
            },
        }

        const log = logger.startCompletion({
            request,
            provider: PROVIDER_IDENTIFIER,
            serverEndpoint: this.serverEndpoint,
        })

        const response = await fetch(this.serverEndpoint, {
            method: 'POST',
            body: JSON.stringify(request),
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${this.accessToken}`,
            },
            signal: abortSignal,
        })

        try {
            const data = (await response.json()) as { generated_text: string }[] | { error: string }

            if ('error' in data) {
                throw new Error(data.error)
            }

            const completions: string[] = data.map((c: { generated_text: string }) =>
                c.generated_text.replace(STOP_WORD, '')
            )
            log?.onComplete(completions)

            return completions.map(content => ({
                prefix: this.prefix,
                content,
            }))
        } catch (error) {
            if (!isAbortError(error)) {
                log?.onError(error)
            }

            throw error
        }
    }
}

export function createProviderConfig(unstableHuggingFaceOptions: UnstableHuggingFaceOptions): ProviderConfig {
    const contextWindowChars = 8_000 // ~ 2k token limit
    return {
        create(options: ProviderOptions) {
            return new UnstableHuggingFaceProvider(options, unstableHuggingFaceOptions)
        },
        maximumContextCharacters: contextWindowChars,
        identifier: PROVIDER_IDENTIFIER,
    }
}
