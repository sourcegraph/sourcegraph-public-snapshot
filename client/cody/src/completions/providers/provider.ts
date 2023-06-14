import { Completion } from '..'
import { ReferenceSnippet } from '../context'

export interface ProviderOptions {
    prefix: string
    suffix: string
    fileName: string
    languageId: string
    snippets: ReferenceSnippet[]
    multilineMode: null | 'block'
    n: number
}

export abstract class Provider {
    protected prefix: string
    protected suffix: string
    protected fileName: string
    protected languageId: string
    protected snippets: ReferenceSnippet[]
    protected multilineMode: null | 'block'
    protected n: number

    constructor({ prefix, suffix, fileName, languageId, snippets, multilineMode, n = 1 }: ProviderOptions) {
        this.prefix = prefix
        this.suffix = suffix
        this.fileName = fileName
        this.languageId = languageId
        this.snippets = snippets
        this.multilineMode = multilineMode
        this.n = n
    }

    /**
     * A rough estimation of how many characters we can add as context to a
     * given prefix/suffix. This is used to early-terminate the context fetching
     * logic and should be a safe upper bound. It can over-fetch and the
     * implementor of the class can decide how many concrete snippets to include
     *
     * Defaults to the full token limit.
     */
    public static estimateContextChars(): number {
        return 2500
    }

    public abstract generateCompletions(abortSignal: AbortSignal): Promise<Completion[]>
}
