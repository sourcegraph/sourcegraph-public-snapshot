import { NEVER } from 'rxjs'

import { isErrorLike } from '@sourcegraph/common'

import { createExtensionHostAPI } from '../api/extension/extensionHostApi'
import { createExtensionHostState } from '../api/extension/extensionHostState'
import { pretendRemote } from '../api/util'
import { PlatformContext } from '../platform/context'

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
                if (
                    settingsCascade.final === null ||
                    settingsCascade.subjects === null ||
                    isErrorLike(settingsCascade.final)
                ) {
                    reject(new Error('Settings are not valid'))
                }

                const extensionHostState = createExtensionHostState(
                    {
                        clientApplication: 'sourcegraph',
                        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-explicit-any
                        initialSettings: settingsCascade as any,
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
