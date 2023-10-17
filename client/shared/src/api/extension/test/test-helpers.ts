import type { Remote } from 'comlink'
import { BehaviorSubject } from 'rxjs'

import type { ClientAPI } from '../../client/api/api'
import type { FlatExtensionHostAPI } from '../../contract'
import { pretendRemote } from '../../util'
import { setActiveLoggers } from '../api/logging'
import { createExtensionAPIFactory } from '../extensionApi'
import type { InitData } from '../extensionHost'
import { createExtensionHostAPI } from '../extensionHostApi'
import { createExtensionHostState } from '../extensionHostState'

export function initializeExtensionHostTest(
    initData: InitData,
    mockMainThreadAPI: Remote<ClientAPI> = pretendRemote<ClientAPI>({}),
    extensionID: string = 'TEST'
): { extensionHostAPI: FlatExtensionHostAPI; extensionAPI: ReturnType<ReturnType<typeof createExtensionAPIFactory>> } {
    // Since the mock main thread API is in the same thread and a connection is synchronously established,
    // we can mock `mainThreadInitializations` as well.
    const mainThreadInitializations = new BehaviorSubject(true)

    const extensionHostState = createExtensionHostState(initData, mockMainThreadAPI, mainThreadInitializations)

    const extensionHostAPI = createExtensionHostAPI(extensionHostState)
    const extensionAPIFactory = createExtensionAPIFactory(extensionHostState, mockMainThreadAPI, initData)
    const extensionAPI = extensionAPIFactory(extensionID)

    setActiveLoggers(extensionHostState)

    return {
        extensionHostAPI,
        extensionAPI,
    }
}
