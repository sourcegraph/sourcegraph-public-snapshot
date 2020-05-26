import { ProxyMarked } from 'comlink'
import { InitData } from '../extensionHost'
import { ExtDocumentsAPI } from './documents'
import { ExtExtensionsAPI } from './extensions'
import { ExtWorkspaceAPI } from './workspace'
import { ExtWindowsAPI } from './windows'
import { FlatExtHostAPI } from '../../contract'

export type ExtensionHostAPIFactory = (initData: InitData) => ExtensionHostAPI

export interface ExtensionHostAPI extends ProxyMarked, FlatExtHostAPI {
    ping(): 'pong'

    documents: ExtDocumentsAPI
    extensions: ExtExtensionsAPI
    workspace: ExtWorkspaceAPI
    windows: ExtWindowsAPI
}
