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
        const fileExtension = this.fileName.split('.').pop()
        const params = {
            debug_ext_path: fileExtension,
            // TODO: Are there unsupported languages?
            lang_prefix: `<|${this.languageId}|>`,
            prefix: this.prefix,
            suffix: this.suffix,
            top_p: 0.95,
            temperature: 0.2,
            max_tokens: 64,
            batch_size: Math.max(this.n, 4),
            // TODO: Figure out the exact format to attach context
            context: '',
            // TODO: Can we pass a stop signal? E.g. to terminate early when we
            // only need single-line completions?
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

        return data.completions.map(c => ({
            prefix: this.prefix,
            content: c.completion,
        }))
    }
}

export function createProviderConfig(unstableCodeGenOptions: UnstableCodeGenOptions): ProviderConfig {
    return {
        create(options: ProviderOptions) {
            return new UnstableCodeGenProvider(options, unstableCodeGenOptions)
        },
        maximumContextCharacters: unstableCodeGenOptions.contextWindowChars,
    }
}
