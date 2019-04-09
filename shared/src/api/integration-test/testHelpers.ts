import 'message-port-polyfill'

import { BehaviorSubject, from, NEVER, NextObserver, Subscribable, throwError } from 'rxjs'
import { first, take } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { EndpointPair, PlatformContext } from '../../platform/context'
import { isDefined } from '../../util/types'
import { ExtensionHostClient } from '../client/client'
import { createExtensionHostClientConnection } from '../client/connection'
import { Model } from '../client/model'
import { Services } from '../client/services'
import { WorkspaceRootWithMetadata } from '../client/services/workspaceService'
import { InitData, startExtensionHost } from '../extension/extensionHost'

interface TestInitData extends Model {
    roots: readonly WorkspaceRootWithMetadata[]
}

const FIXTURE_INIT_DATA: TestInitData = {
    roots: [{ uri: 'file:///' }],
    visibleViewComponents: [
        {
            type: 'CodeEditor',
            item: {
                uri: 'file:///f',
                languageId: 'l',
                text: 't',
            },
            selections: [],
            isActive: true,
        },
    ],
}

interface TestContext {
    client: ExtensionHostClient
    extensionAPI: typeof sourcegraph
}

interface Mocks
    extends Pick<
        PlatformContext,
        | 'settings'
        | 'updateSettings'
        | 'queryGraphQL'
        | 'getScriptURLForExtension'
        | 'clientApplication'
        | 'sideloadedExtensionURL'
    > {}

const NOOP_MOCKS: Mocks = {
    settings: NEVER,
    updateSettings: () => Promise.reject(new Error('Mocks#updateSettings not implemented')),
    queryGraphQL: () => throwError(new Error('Mocks#queryGraphQL not implemented')),
    getScriptURLForExtension: scriptURL => scriptURL,
    clientApplication: 'sourcegraph',
    sideloadedExtensionURL: new BehaviorSubject<string | null>(null),
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
    TestContext & {
        model: Subscribable<Model> & { value: Model } & NextObserver<Model>
        services: Services
    }
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

    const services = new Services(mocks)
    const initData: InitData = {
        sourcegraphURL: 'https://example.com/',
        clientApplication: 'sourcegraph',
    }
    const client = await createExtensionHostClientConnection(clientEndpoints, services, initData)

    const extensionAPI = await extensionHost.extensionAPI
    services.model.model.next(initModel)
    services.workspace.roots.next(initModel.roots)

    // Wait for initModel to be initialized
    await Promise.all([
        from(extensionAPI.workspace.openedTextDocuments)
            .pipe(take((initModel.visibleViewComponents || []).length))
            .toPromise(),
        from(extensionAPI.app.activeWindowChanges)
            .pipe(first(isDefined))
            .toPromise(),
    ])

    return {
        client,
        extensionAPI,
        services,
        model: services.model.model,
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
