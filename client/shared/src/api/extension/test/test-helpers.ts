import { Remote } from 'comlink'

import { FlatExtensionHostAPI, MainThreadAPI } from '../../contract'
import { pretendRemote } from '../../util'
import { createExtensionAPI } from '../extensionApi'
import { InitData } from '../extensionHost'
import { createExtensionHostAPI } from '../extensionHostApi'
import { createExtensionHostState } from '../extensionHostState'

export function initializeExtensionHostTest(
    initData: Pick<InitData, 'initialSettings' | 'clientApplication'>,
    mockMainThreadAPI: Remote<MainThreadAPI> = pretendRemote<MainThreadAPI>({})
): { extensionHostAPI: FlatExtensionHostAPI; extensionAPI: ReturnType<typeof createExtensionAPI> } {
    const extensionHostState = createExtensionHostState(initData, mockMainThreadAPI)

    const extensionHostAPI = createExtensionHostAPI(extensionHostState)
    const extensionAPI = createExtensionAPI(extensionHostState, mockMainThreadAPI)

    return {
        extensionHostAPI,
        extensionAPI,
    }
}
