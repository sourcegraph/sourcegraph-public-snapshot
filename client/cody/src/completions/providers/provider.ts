import { Completion } from '..'
import { ReferenceSnippet } from '../context'

export interface ProviderConfig {
    /**
     * A factory to create instances of the provider. This pattern allows us to inject provider
     * specific parameters outside of the callers of the factory.
     */
    create(options: ProviderOptions): Provider

    /**
     * The maximum number of unicode characters that should be included in the
     * context window. Note that these are not tokens as the definition can vary
     * between models.
     *
     * This value is used for determining the length of prefix, suffix, and
     * snippets and can be validated by the provider implementing it.
     */
    maximumContextCharacters: number

    /**
     * A string identifier used in event logs
     */
    identifier: string
}

export interface ProviderOptions {
    prefix: string
    suffix: string
    fileName: string
    languageId: string
    snippets: ReferenceSnippet[]
    multilineMode: null | 'block'
    // Relative length to `maximumContextCharacters`
    responsePercentage: number
    prefixPercentage: number
    suffixPercentage: number
    // Number of parallel LLM requests per completion.
    n: number
}

export abstract class Provider {
    protected prefix: string
    protected suffix: string
    protected fileName: string
    protected languageId: string
    protected snippets: ReferenceSnippet[]
    protected multilineMode: null | 'block'
    protected responsePercentage: number
    protected prefixPercentage: number
    protected suffixPercentage: number
    protected n: number

    constructor({
        prefix,
        suffix,
        fileName,
        languageId,
        snippets,
        multilineMode,
        responsePercentage,
        prefixPercentage,
        suffixPercentage,
        n = 1,
    }: ProviderOptions) {
        this.prefix = prefix
        this.suffix = suffix
        this.fileName = fileName
        this.languageId = languageId
        this.snippets = snippets
        this.multilineMode = multilineMode
        this.responsePercentage = responsePercentage
        this.prefixPercentage = prefixPercentage
        this.suffixPercentage = suffixPercentage
        this.n = n
    }

    public abstract generateCompletions(abortSignal: AbortSignal): Promise<Completion[]>
}
