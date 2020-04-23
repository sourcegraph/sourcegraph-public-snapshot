import { ProxyMarked } from '@sourcegraph/comlink'
import { InitData } from '../extensionHost'
import { ExtConfigurationAPI } from './configuration'
import { ExtDocumentsAPI } from './documents'
import { ExtExtensionsAPI } from './extensions'
import { ExtRootsAPI } from './roots'
import { ExtWindowsAPI } from './windows'

export type ExtensionHostAPIFactory = (initData: InitData) => ExtensionHostAPI

export interface ExtensionHostAPI extends ProxyMarked {
    ping(): 'pong'

    documents: ExtDocumentsAPI
    extensions: ExtExtensionsAPI
    roots: ExtRootsAPI
    windows: ExtWindowsAPI
    configuration: ExtConfigurationAPI<any>
}
