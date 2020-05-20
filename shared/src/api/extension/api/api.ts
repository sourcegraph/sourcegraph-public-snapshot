import { ProxyMarked } from 'comlink'
import { InitData } from '../extensionHost'
import { ExtConfigurationAPI } from './configuration'
import { ExtDocumentsAPI } from './documents'
import { ExtExtensionsAPI } from './extensions'
import { ExtWorkspaceAPI } from './workspace'
import { ExtWindowsAPI } from './windows'

export type ExtensionHostAPIFactory = (initData: InitData) => ExtensionHostAPI

export interface ExtensionHostAPI extends ProxyMarked {
    ping(): 'pong'

    documents: ExtDocumentsAPI
    extensions: ExtExtensionsAPI
    workspace: ExtWorkspaceAPI
    windows: ExtWindowsAPI
    configuration: ExtConfigurationAPI<any>
}
