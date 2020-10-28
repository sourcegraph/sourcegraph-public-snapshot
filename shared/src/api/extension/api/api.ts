import { ProxyMarked } from 'comlink'
import { InitData } from '../extensionHost'
import { ExtensionWindowsAPI } from './windows'
import { FlatExtensionHostAPI } from '../../contract'

export type ExtensionHostAPIFactory = (initData: InitData) => ExtensionHostAPI

export interface ExtensionHostAPI extends ProxyMarked, FlatExtensionHostAPI {
    ping(): 'pong'

    windows: ExtensionWindowsAPI
}
