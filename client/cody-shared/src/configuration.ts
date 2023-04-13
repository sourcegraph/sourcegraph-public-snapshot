export type ConfigurationUseContext = 'embeddings' | 'keyword' | 'none' | 'blended'

export interface Configuration {
    enabled: boolean
    serverEndpoint: string
    codebase?: string
    debug: boolean
    useContext: ConfigurationUseContext
    experimentalSuggest: boolean
    anthropicKey: string | null
    customHeaders: Record<string, string>
}
