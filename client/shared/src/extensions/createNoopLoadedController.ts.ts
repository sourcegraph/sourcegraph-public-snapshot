import { Remote } from 'comlink'
import { from, Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'

import { ExposedToClient, initMainThreadAPI } from '../api/client/mainthread-api'
import { FlatExtensionHostAPI } from '../api/contract'
import { proxySubscribable } from '../api/extension/api/common'
import { createExtensionHostAPI } from '../api/extension/extensionHostApi'
import { createExtensionHostState } from '../api/extension/extensionHostState'
import { pretendRemote, syncPromiseSubscription } from '../api/util'
import { newCodeIntelAPI } from '../codeintel/api'
import { CodeIntelContext, newSettingsGetter } from '../codeintel/legacy-extensions/api'
import { PlatformContext } from '../platform/context'
import { isSettingsValid } from '../settings/settings'

import { Controller } from './controller'

export function createNoopController(platformContext: PlatformContext): Controller {
    const api: Promise<{
        remoteExtensionHostAPI: Remote<FlatExtensionHostAPI>
        exposedToClient: ExposedToClient
    }> = new Promise((resolve, reject) => {
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
            const extensionHostAPI = injectNewCodeintel(createExtensionHostAPI(extensionHostState), {
                requestGraphQL: platformContext.requestGraphQL,
                telemetryService: platformContext.telemetryService,
                settings: newSettingsGetter(platformContext.settings),
            })
            const remoteExtensionHostAPI = pretendRemote(extensionHostAPI)
            const exposedToClient = initMainThreadAPI(remoteExtensionHostAPI, platformContext).exposedToClient
            // We don't have to load any extensions so we are already done
            extensionHostState.haveInitialExtensionsLoaded.next(true)

            resolve({ remoteExtensionHostAPI, exposedToClient })
        })
    })
    return {
        executeCommand: (parameters, suppressNotificationOnError) =>
            api.then(({ exposedToClient }) => exposedToClient.executeCommand(parameters, suppressNotificationOnError)),
        commandErrors: from(api).pipe(switchMap(({ exposedToClient }) => exposedToClient.commandErrors)),
        registerCommand: entryToRegister =>
            syncPromiseSubscription(
                api.then(({ exposedToClient }) => exposedToClient.registerCommand(entryToRegister))
            ),
        extHostAPI: api.then(({ remoteExtensionHostAPI }) => remoteExtensionHostAPI),
        unsubscribe: () => {},
    }
}

// Replaces codeintel functions from the "old" extension/webworker extension API
// with new implementations of code that lives in this repository. The old
// implementation invoked codeintel functions via webworkers, and the codeintel
// implementation lived in a separate repository
// https://github.com/sourcegraph/code-intel-extensions Ideally, we should
// update all the usages of `comlink.Remote<FlatExtensionHostAPI>` with the new
// `CodeIntelAPI` interfaces, but that would require refactoring a lot of files.
// To minimize the risk of breaking changes caused by the deprecation of
// extensions, we monkey patch the old implementation with new implementations.
// The benefit of monkey patching is that we can optionally disable if for
// customers that choose to enable the legacy extensions.
export function injectNewCodeintel(
    old: FlatExtensionHostAPI,
    codeintelContext: CodeIntelContext
): FlatExtensionHostAPI {
    const codeintel = newCodeIntelAPI(codeintelContext)
    function thenMaybeLoadingResult<T>(promise: Observable<T>): Observable<MaybeLoadingResult<T>> {
        return promise.pipe(
            map(result => {
                const maybeLoadingResult: MaybeLoadingResult<T> = { isLoading: false, result }
                return maybeLoadingResult
            })
        )
    }

    const codeintelOverrides: Pick<
        FlatExtensionHostAPI,
        | 'getHover'
        | 'getDocumentHighlights'
        | 'getReferences'
        | 'getDefinition'
        | 'getLocations'
        | 'hasReferenceProvidersForDocument'
    > = {
        hasReferenceProvidersForDocument(textParameters) {
            return proxySubscribable(codeintel.hasReferenceProvidersForDocument(textParameters))
        },
        getLocations(id, parameters) {
            console.log({ id })
            return proxySubscribable(thenMaybeLoadingResult(codeintel.getImplementations(parameters)))
        },
        getDefinition(parameters) {
            return proxySubscribable(thenMaybeLoadingResult(codeintel.getDefinition(parameters)))
        },
        getReferences(parameters, context) {
            return proxySubscribable(thenMaybeLoadingResult(codeintel.getReferences(parameters, context)))
        },
        getDocumentHighlights: (textParameters: TextDocumentPositionParameters) =>
            proxySubscribable(codeintel.getDocumentHighlights(textParameters)),
        getHover: (textParameters: TextDocumentPositionParameters) =>
            proxySubscribable(thenMaybeLoadingResult(codeintel.getHover(textParameters))),
    }

    return new Proxy(old, {
        get(target, prop) {
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-explicit-any
            const codeintelFunction = (codeintelOverrides as any)[prop]
            if (codeintelFunction) {
                return codeintelFunction
            }
            // eslint-disable-next-line prefer-rest-params
            return Reflect.get(target, prop, ...arguments)
        },
    })
}
