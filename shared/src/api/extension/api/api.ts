import { ProxyMarked } from 'comlink'
import { InitData } from '../extensionHost'
import { ExtensionWindowsAPI } from './windows'
import { FlatExtHostAPI } from '../../contract'

export type ExtensionHostAPIFactory = (initData: InitData) => ExtensionHostAPI

export interface ExtensionHostAPI extends ProxyMarked, FlatExtHostAPI {
    ping(): 'pong'

    windows: ExtensionWindowsAPI
}
