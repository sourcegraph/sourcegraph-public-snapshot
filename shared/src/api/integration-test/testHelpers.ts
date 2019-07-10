import 'message-port-polyfill'

import { BehaviorSubject, from, NEVER, throwError } from 'rxjs'
import { filter, first, switchMap, take } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { EndpointPair, PlatformContext } from '../../platform/context'
import { isDefined } from '../../util/types'
import { ExtensionHostClient } from '../client/client'
import { createExtensionHostClientConnection } from '../client/connection'
import { Services } from '../client/services'
import { CodeEditor } from '../client/services/editorService'
import { TextModel } from '../client/services/modelService'
import { WorkspaceRootWithMetadata } from '../client/services/workspaceService'
import { InitData, startExtensionHost } from '../extension/extensionHost'

export function assertToJSON(a: any, expected: any): void {
    const raw = JSON.stringify(a)
    const actual = JSON.parse(raw)
    expect(actual).toEqual(expected)
}

interface TestInitData {
    roots: readonly WorkspaceRootWithMetadata[]
    models?: readonly TextModel[]
    editors: readonly Pick<CodeEditor, Exclude<keyof CodeEditor, 'editorId'>>[]
}

const FIXTURE_INIT_DATA: TestInitData = {
    roots: [{ uri: 'file:///' }],
    models: [{ uri: 'file:///f', text: 't', languageId: 'l' }],
    editors: [
        {
            type: 'CodeEditor',
            resource: 'file:///f',
            selections: [],
            isActive: true,
        },
    ],
}

interface Mocks
    extends Pick<
        PlatformContext,
        | 'settings'
        | 'updateSettings'
        | 'requestGraphQL'
        | 'getScriptURLForExtension'
        | 'clientApplication'
        | 'sideloadedExtensionURL'
    > {}

const NOOP_MOCKS: Mocks = {
    settings: NEVER,
    updateSettings: () => Promise.reject(new Error('Mocks#updateSettings not implemented')),
    requestGraphQL: () => throwError(new Error('Mocks#queryGraphQL not implemented')),
    getScriptURLForExtension: scriptURL => scriptURL,
    clientApplication: 'sourcegraph',
    sideloadedExtensionURL: new BehaviorSubject<string | null>(null),
}

/**requestGraphQL
 * Set up a new client-extension integration test.
 *
 * @internal
 */
export async function integrationTestContext(
    partialMocks: Partial<Mocks> = NOOP_MOCKS,
    initModel: TestInitData = FIXTURE_INIT_DATA
): Promise<{
    client: ExtensionHostClient
    extensionAPI: typeof sourcegraph
    services: Services
}> {
    const mocks = partialMocks ? { ...NOOP_MOCKS, ...partialMocks } : NOOP_MOCKS

    const clientAPIChannel = new MessageChannel()
    const extensionHostAPIChannel = new MessageChannel()
    const extensionHostEndpoints: EndpointPair = {
        proxy: clientAPIChannel.port2,
        expose: extensionHostAPIChannel.port2,
    }
    const clientEndpoints: EndpointPair = {
        proxy: extensionHostAPIChannel.port1,
        expose: clientAPIChannel.port1,
    }

    const extensionHost = startExtensionHost(extensionHostEndpoints)

    const services = new Services(mocks)
    const initData: InitData = {
        sourcegraphURL: 'https://example.com/',
        clientApplication: 'sourcegraph',
    }
    const client = await createExtensionHostClientConnection(clientEndpoints, services, initData)

    const extensionAPI = await extensionHost.extensionAPI
    if (initModel.models) {
        for (const model of initModel.models) {
            services.model.addModel(model)
        }
    }
    for (const editor of initModel.editors) {
        services.editor.addEditor(editor)
    }
    services.workspace.roots.next(initModel.roots)

    // Wait for initModel to be initialized
    await Promise.all([
        from(extensionAPI.workspace.openedTextDocuments)
            .pipe(take(initModel.editors.length))
            .toPromise(),
        from(extensionAPI.app.activeWindowChanges)
            .pipe(
                first(isDefined),
                switchMap(activeWindow =>
                    from(activeWindow.activeViewComponentChanges).pipe(
                        filter(isDefined),
                        take(initModel.editors.length)
                    )
                )
            )
            .toPromise(),
    ])

    return {
        client,
        extensionAPI,
        services,
    }
}

/**
 * Returns a {@link Promise} and a function. The {@link Promise} blocks until the returned function is called.
 *
 * @internal
 */
export function createBarrier(): { wait: Promise<void>; done: () => void } {
    let done!: () => void
    const wait = new Promise<void>(resolve => (done = resolve))
    return { wait, done }
}

export function collectSubscribableValues<T>(subscribable: sourcegraph.Subscribable<T>): T[] {
    const values: T[] = []
    subscribable.subscribe(value => values.push(value))
    return values
}
