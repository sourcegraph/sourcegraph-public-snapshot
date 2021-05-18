import { Remote } from 'comlink'

import { ClientAPI } from '../../client/api/api'
import { FlatExtensionHostAPI } from '../../contract'
import { pretendRemote } from '../../util'
import { setActiveLoggers } from '../api/logging'
import { createExtensionAPIFactory } from '../extensionApi'
import { InitData } from '../extensionHost'
import { createExtensionHostAPI } from '../extensionHostApi'
import { createExtensionHostState } from '../extensionHostState'

export function initializeExtensionHostTest(
    initData: InitData,
    mockMainThreadAPI: Remote<ClientAPI> = pretendRemote<ClientAPI>({}),
    extensionID: string = 'TEST'
): { extensionHostAPI: FlatExtensionHostAPI; extensionAPI: ReturnType<ReturnType<typeof createExtensionAPIFactory>> } {
    const extensionHostState = createExtensionHostState(initData, mockMainThreadAPI)

    const extensionHostAPI = createExtensionHostAPI(extensionHostState)
    const extensionAPIFactory = createExtensionAPIFactory(extensionHostState, mockMainThreadAPI, initData)
    const extensionAPI = extensionAPIFactory(extensionID)

    setActiveLoggers(extensionHostState)

    return {
        extensionHostAPI,
        extensionAPI,
    }
}
