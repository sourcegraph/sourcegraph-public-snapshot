import { NEVER } from 'rxjs'

import { createExtensionHostAPI } from '../api/extension/extensionHostApi'
import { createExtensionHostState } from '../api/extension/extensionHostState'
import { pretendRemote } from '../api/util'
import { PlatformContext } from '../platform/context'
import { isSettingsValid } from '../settings/settings'

import { Controller } from './controller'

export function createNoopController(platformContext: PlatformContext): Controller {
    return {
        executeCommand: () => Promise.resolve(),
        commandErrors: NEVER,
        registerCommand: () => ({
            unsubscribe: () => {},
        }),
        extHostAPI: new Promise((resolve, reject) => {
            platformContext.settings.subscribe(settingsCascade => {
                if (!isSettingsValid(settingsCascade)) {
                    reject(new Error('Settings are not valid'))
                    return
                }

                const extensionHostState = createExtensionHostState(
                    {
                        clientApplication: 'sourcegraph',
                        initialSettings: settingsCascade,
                    },
                    null,
                    null
                )
                const extensionHostAPI = pretendRemote(createExtensionHostAPI(extensionHostState))

                resolve(extensionHostAPI)
            })
        }),

        unsubscribe: () => {},
    }
}
