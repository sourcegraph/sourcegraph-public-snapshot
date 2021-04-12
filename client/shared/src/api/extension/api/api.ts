import { ProxyMarked } from 'comlink'

import { FlatExtensionHostAPI } from '../../contract'
import { InitData } from '../extensionHost'

export type ExtensionHostAPIFactory = (initData: InitData) => ExtensionHostAPI

export interface ExtensionHostAPI extends ProxyMarked, FlatExtensionHostAPI {
    ping(): 'pong'
}
