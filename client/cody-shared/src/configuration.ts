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
    experimentalSuggest: boolean
    experimentalChatPredictions: boolean
    experimentalInline: boolean
    experimentalGuardrails: boolean
    experimentalNonStop: boolean
    completionsAdvancedProvider: 'anthropic' | 'unstable-codegen' | 'unstable-huggingface'
    completionsAdvancedServerEndpoint: string | null
    completionsAdvancedAccessToken: string | null
    completionsAdvancedCache: boolean
    completionsAdvancedEmbeddings: boolean
}

export interface ConfigurationWithAccessToken extends Configuration {
    /** The access token, which is stored in the secret storage (not configuration). */
    accessToken: string | null
}
