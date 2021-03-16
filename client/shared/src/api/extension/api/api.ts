import { ProxyMarked } from 'comlink'
import { InitData } from '../extensionHost'
import { FlatExtensionHostAPI } from '../../contract'

export type ExtensionHostAPIFactory = (initData: InitData) => ExtensionHostAPI

export interface ExtensionHostAPI extends ProxyMarked, FlatExtensionHostAPI {
    ping(): 'pong'
}
