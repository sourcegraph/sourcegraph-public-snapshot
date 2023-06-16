export type ConfigurationUseContext = 'embeddings' | 'keyword' | 'none' | 'blended' | 'unified'

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
    completionsAdvancedProvider: 'anthropic' | 'unstable-codegen'
    completionsAdvancedServerEndpoint: string | null
}

export interface ConfigurationWithAccessToken extends Configuration {
    /** The access token, which is stored in the secret storage (not configuration). */
    accessToken: string | null
}
