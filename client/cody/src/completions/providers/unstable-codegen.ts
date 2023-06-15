import { Completion } from '..'

import { Provider, ProviderConfig, ProviderOptions } from './provider'

interface UnstableCodeGenOptions {
    serverEndpoint: string
    contextWindowChars: number
}

export class UnstableCodeGenProvider extends Provider {
    private serverEndpoint: string

    constructor(options: ProviderOptions, unstableCodeGenOptions: UnstableCodeGenOptions) {
        super(options)
        this.serverEndpoint = unstableCodeGenOptions.serverEndpoint
    }

    public async generateCompletions(abortSignal: AbortSignal): Promise<Completion[]> {
        const params = {
            debug_ext_path: 'cody',
            // TODO: Are there unsupported languages?
            lang_prefix: `<|${this.languageId}|>`,
            prefix: this.prefix,
            suffix: this.suffix,
            top_p: 0.95,
            temperature: 0.2,
            max_tokens: this.multilineMode === null ? 40 : 128,
            batch_size: 4, // this.n,
            // TODO: Figure out the exact format to attach context
            context: '',
        }

        const response = await fetch(this.serverEndpoint, {
            method: 'POST',
            body: JSON.stringify(params),
            headers: {
                'Content-Type': 'application/json',
            },
            signal: abortSignal,
        })

        const data = (await response.json()) as { completions: { completion: string }[] }

        console.log(params, data, [...response.headers.entries()])

        return data.completions.map(c => ({
            prefix: this.prefix,
            content: c.completion,
        }))
    }
}

export function createProviderConfig(
    unstableCodeGenOptions: Omit<UnstableCodeGenOptions, 'contextWindowChars'>
): ProviderConfig {
    const contextWindowChars = 8_000 // ~ 2k token limit
    return {
        create(options: ProviderOptions) {
            return new UnstableCodeGenProvider(options, { ...unstableCodeGenOptions, contextWindowChars })
        },
        maximumContextCharacters: contextWindowChars,
        identifier: 'codegen',
    }
}
