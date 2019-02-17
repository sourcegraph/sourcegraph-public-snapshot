import { ExtConfigurationAPI } from './configuration'
import { ExtDocumentsAPI } from './documents'
import { ExtExtensionsAPI } from './extensions'
import { ExtRootsAPI } from './roots'
import { ExtWindowsAPI } from './windows'

export interface ExtensionHostAPI {
    ping(): 'pong'
    documents: ExtDocumentsAPI
    extensions: ExtExtensionsAPI
    roots: ExtRootsAPI
    windows: ExtWindowsAPI
    configuration: ExtConfigurationAPI<any>
}
