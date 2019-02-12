import { ProxyValue } from 'comlink'
import { ExtCommandsAPI } from './commands'
import { ExtConfigurationAPI } from './configuration'
import { ExtDocumentsAPI } from './documents'
import { ExtExtensionsAPI } from './extensions'
import { ExtLanguageFeaturesAPI } from './languageFeatures'
import { ExtRootsAPI } from './roots'
import { ExtWindowsAPI } from './windows'

export interface ExtensionHostAPI {
    documents: ExtDocumentsAPI & ProxyValue
    extensions: ExtExtensionsAPI & ProxyValue
    roots: ExtRootsAPI & ProxyValue
    windows: ExtWindowsAPI & ProxyValue
    configuration: ExtConfigurationAPI<any> & ProxyValue
    languageFeatures: ExtLanguageFeaturesAPI & ProxyValue
    commands: ExtCommandsAPI & ProxyValue
}
