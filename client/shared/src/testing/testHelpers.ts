import 'message-port-polyfill'

import { Remote } from 'comlink'
import { throwError, of, Subscription, Unsubscribable, Subscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

import { createExtensionHostClientConnection } from '../api/client/connection'
import { ExposedToClient } from '../api/client/mainthread-api'
import { FlatExtensionHostAPI, MainThreadAPI } from '../api/contract'
import { InitData, startExtensionHost } from '../api/extension/extensionHost'
import { WorkspaceRootWithMetadata } from '../api/extension/extensionHostApi'
import { TextDocumentData, ViewerData } from '../api/viewerTypes'
import { EndpointPair, PlatformContext } from '../platform/context'

export function assertToJSON(a: any, expected: any): void {
    const raw = JSON.stringify(a)
    const actual = JSON.parse(raw)
    expect(actual).toEqual(expected)
}

interface TestInitData {
    roots: readonly WorkspaceRootWithMetadata[]
    textDocuments?: readonly TextDocumentData[]
    viewers: readonly ViewerData[]
}

const FIXTURE_INIT_DATA: TestInitData = {
    roots: [{ uri: 'file:///' }],
    textDocuments: [{ uri: 'file:///f', text: 't', languageId: 'l' }],
    viewers: [
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
        | 'getGraphQLClient'
        | 'requestGraphQL'
        | 'clientApplication'
        | 'showMessage'
        | 'showInputBox'
    > {}

const NOOP_MOCKS: Mocks = {
    settings: of({ final: {}, subjects: [] }),
    updateSettings: () => Promise.reject(new Error('Mocks#updateSettings not implemented')),
    getGraphQLClient: () => Promise.reject(new Error('Mocks#getGraphQLClient not implemented')),
    requestGraphQL: () => throwError(new Error('Mocks#queryGraphQL not implemented')),
    clientApplication: 'sourcegraph',
}

/**
 * Set up a new client-extension integration test.
 *
 * @internal
 */
export async function integrationTestContext(
    partialMocks: Partial<Mocks> = NOOP_MOCKS,
    initModel: TestInitData = FIXTURE_INIT_DATA
): Promise<
    {
        extensionAPI: typeof sourcegraph
        extensionHostAPI: Remote<FlatExtensionHostAPI>
        mainThreadAPI: MainThreadAPI
        exposedToClient: ExposedToClient
    } & Unsubscribable
> {
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

    const initData: Omit<InitData, 'initialSettings'> = {
        sourcegraphURL: 'https://example.com/',
        clientApplication: 'sourcegraph',
    }

    const {
        api: extensionHostAPI,
        mainThreadAPI,
        exposedToClient,
    } = await createExtensionHostClientConnection(
        Promise.resolve({
            endpoints: clientEndpoints,
            subscription: new Subscription(),
        }),
        initData,
        mocks
    )

    const extensionAPI = await extensionHost.extensionAPI

    await Promise.all((initModel.textDocuments || []).map(model => extensionHostAPI.addTextDocumentIfNotExists(model)))
    await Promise.all(initModel.viewers.map(viewer => extensionHostAPI.addViewerIfNotExists(viewer)))
    await Promise.all(initModel.roots.map(root => extensionHostAPI.addWorkspaceRoot(root)))

    return {
        extensionAPI,
        extensionHostAPI,
        mainThreadAPI,
        exposedToClient,
        unsubscribe: () => extensionHost.unsubscribe(),
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

export function collectSubscribableValues<T>(subscribable: Subscribable<T>): T[] {
    const values: T[] = []
    subscribable.subscribe(value => values.push(value))
    return values
}
