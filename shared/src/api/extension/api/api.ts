import { ProxyMarked } from 'comlink'
import { InitData } from '../extensionHost'
import { ExtDocumentsAPI } from './documents'
import { ExtExtensionsAPI } from './extensions'
import { ExtensionWindowsAPI } from './windows'
import { FlatExtHostAPI } from '../../contract'

export type ExtensionHostAPIFactory = (initData: InitData) => ExtensionHostAPI

export interface ExtensionHostAPI extends ProxyMarked, FlatExtHostAPI {
    ping(): 'pong'

    documents: ExtDocumentsAPI
    extensions: ExtExtensionsAPI
    windows: ExtensionWindowsAPI
}
