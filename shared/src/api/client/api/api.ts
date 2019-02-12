import { ProxyValue } from 'comlink'
import { ClientConfigurationAPI } from './configuration'
import { ClientLanguageFeaturesAPI } from './languageFeatures'
import { ClientSearchAPI } from './search'

/**
 * The API that is exposed from the client (main thread) to the extension host (worker)
 */
export interface ClientAPI {
    ping(): 'pong'

    configuration: ClientConfigurationAPI & ProxyValue
    search: ClientSearchAPI & ProxyValue
    languageFeatures: ClientLanguageFeaturesAPI & ProxyValue
}
