import { Completion } from '..'
// import { logger } from '../../log'
import { ReferenceSnippet } from '../context'
import { isAbortError } from '../utils'

import { Provider, ProviderConfig, ProviderOptions } from './provider'

interface UnstableCodeGenOptions {
    serverEndpoint: string
}

const PROVIDER_IDENTIFIER = 'codegen'

export class UnstableCodeGenProvider extends Provider {
    private serverEndpoint: string

    constructor(options: ProviderOptions, unstableCodeGenOptions: UnstableCodeGenOptions) {
        super(options)
        this.serverEndpoint = unstableCodeGenOptions.serverEndpoint
    }

    public async generateCompletions(abortSignal: AbortSignal, snippets: ReferenceSnippet[]): Promise<Completion[]> {
        const params = {
            debug_ext_path: 'cody',
            lang_prefix: `<|${mapVSCodeLanguageIdToModelId(this.languageId)}|>`,
            prefix: this.prefix,
            suffix: this.suffix,
            top_p: 0.95,
            temperature: 0.2,
            max_tokens: this.multilineMode === null ? 40 : 128,
            // The backend expects an even number of requests since it will
            // divide it into two different batches.
            batch_size: makeEven(4),
            // TODO: Figure out the exact format to attach context
            context: JSON.stringify(prepareContext(snippets, this.fileName)),
            completion_type: 'automatic',
        }

        // const log = logger.startCompletion({
        //     params,
        //     provider: PROVIDER_IDENTIFIER,
        //     serverEndpoint: this.serverEndpoint,
        // })
        const response = await fetch(this.serverEndpoint, {
            method: 'POST',
            body: JSON.stringify(params),
            headers: {
                'Content-Type': 'application/json',
            },
            signal: abortSignal,
        })

        try {
            const data = (await response.json()) as { completions: { completion: string }[] }

            const completions: string[] = data.completions.map(c => postProcess(c.completion, this.multilineMode))
            // log?.onComplete(completions)

            return completions.map(content => ({
                prefix: this.prefix,
                content,
            }))
        } catch (error) {
            if (!isAbortError(error)) {
                // log?.onError(error)
            }

            throw error
        }
    }
}

function postProcess(content: string, multilineMode: null | 'block'): string {
    // The model might return multiple lines for single line completions because
    // we are only able to specify a token limit.
    if (multilineMode === null && content.includes('\n')) {
        content = content.slice(0, content.indexOf('\n'))
    }

    return content.trim()
}

// Handles some inconsistencies between the VS Code language ID and the model's
// required language identifier.
function mapVSCodeLanguageIdToModelId(languageId: string): string {
    switch (languageId) {
        case 'typescript':
        case 'typescriptreact':
            return 'typescript'
        case 'javascript':
        case 'javascriptreact':
            return 'javascript'
        case 'css':
        case 'scss':
        case 'sass':
            return 'css'
        case 'c-sharp':
            return 'csharp'
        case 'shellscript':
            return 'shell'
        default:
            return languageId
    }
}

function makeEven(number: number): number {
    if (number % 2 === 1) {
        return number + 1
    }
    return number
}

interface Context {
    current_file_path: string
    windows: {
        file_path: string
        text: string
        similarity: number
    }[]
}

function prepareContext(snippets: ReferenceSnippet[], fileName: string): Context {
    const windows: Context['windows'] = []

    // the model expects a similarly to rank the order and priority to insert
    // snippets. Since we already have ranked results and do not expose the
    // score, we can create an artificial score for simplicity.
    let similarity = 0.5
    for (const snippet of snippets) {
        // Slightly decrease similarity between subsequent windows
        similarity *= 0.99
        windows.push({
            file_path: snippet.fileName,
            text: snippet.content,
            similarity,
        })
    }

    return {
        current_file_path: fileName,
        windows,
    }
}

export function createProviderConfig(unstableCodeGenOptions: UnstableCodeGenOptions): ProviderConfig {
    const contextWindowChars = 8_000 // ~ 2k token limit
    return {
        create(options: ProviderOptions) {
            return new UnstableCodeGenProvider(options, unstableCodeGenOptions)
        },
        maximumContextCharacters: contextWindowChars,
        enableExtendedMultilineTriggers: false,
        identifier: PROVIDER_IDENTIFIER,
    }
}
