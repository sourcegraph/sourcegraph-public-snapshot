export type ConfigurationUseContext = 'embeddings' | 'keyword' | 'none' | 'blended' | 'unified'

// Should we share VS Code specific config via cody-shared?
export interface Configuration {
    serverEndpoint: string
    codebase?: string
    debugEnable: boolean
    debugFilter: RegExp | null
    debugVerbose: boolean
    useContext: ConfigurationUseContext
    customHeaders: Record<string, string>
    autocomplete: boolean
    experimentalChatPredictions: boolean
    experimentalInline: boolean
    experimentalGuardrails: boolean
    experimentalNonStop: boolean
    autocompleteAdvancedProvider: 'anthropic' | 'unstable-codegen' | 'unstable-huggingface'
    autocompleteAdvancedServerEndpoint: string | null
    autocompleteAdvancedAccessToken: string | null
    autocompleteAdvancedCache: boolean
    autocompleteAdvancedEmbeddings: boolean
}

export interface ConfigurationWithAccessToken extends Configuration {
    /** The access token, which is stored in the secret storage (not configuration). */
    accessToken: string | null
}
