import type { ProxyMarked } from 'comlink'

import type { FlatExtensionHostAPI } from '../../contract'
import type { InitData } from '../extensionHost'

export type ExtensionHostAPIFactory = (initData: InitData) => ExtensionHostAPI

export interface ExtensionHostAPI extends ProxyMarked, FlatExtensionHostAPI {
    ping(): 'pong'
    /**
     * Main thread calls this to notify the extension host that `MainThreadAPI` has
     * been created and exposed.
     * */
    mainThreadAPIInitialized: () => void
}
