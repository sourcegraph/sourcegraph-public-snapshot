export type ConfigurationUseContext = 'embeddings' | 'keyword' | 'none' | 'blended'

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
}

export interface ConfigurationWithAccessToken extends Configuration {
    /** The access token, which is stored in the secret storage (not configuration). */
    accessToken: string | null
}
