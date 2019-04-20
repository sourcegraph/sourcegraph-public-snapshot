import 'message-port-polyfill'

import { BehaviorSubject, from, NEVER, throwError } from 'rxjs'
import { filter, first, switchMap, take, tap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { EndpointPair, PlatformContext } from '../../platform/context'
import { isDefined } from '../../util/types'
import { ExtensionHostClient } from '../client/client'
import { createExtensionHostClientConnection } from '../client/connection'
import { Services } from '../client/services'
import { CodeEditor } from '../client/services/editorService'
import { WorkspaceRootWithMetadata } from '../client/services/workspaceService'
import { InitData, startExtensionHost } from '../extension/extensionHost'

interface TestInitData {
    roots: readonly WorkspaceRootWithMetadata[]
    editors: readonly Pick<CodeEditor, Exclude<keyof CodeEditor, 'editorId'>>[]
}

const FIXTURE_INIT_DATA: TestInitData = {
    roots: [{ uri: 'file:///' }],
    editors: [
        {
            type: 'CodeEditor',
            resource: 'file:///f',
            model: {
                uri: 'file:///f',
                languageId: 'l',
                text: 't',
            },
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
): Promise<{
    client: ExtensionHostClient
    extensionAPI: typeof sourcegraph
    services: Services
}> {
    const mocks = partialMocks ? { ...NOOP_MOCKS, ...partialMocks } : NOOP_MOCKS

    const clientAPIChannel = createMessageChannel()
    const extensionHostAPIChannel = createMessageChannel()
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
    for (const { model, ...editor } of initModel.editors) {
        services.model.addModel(model)
        services.editor.addEditor(editor)
    }
    services.workspace.roots.next(initModel.roots)

    // Wait for initModel to be initialized
    await Promise.all([
        // TODO!(sqs) seems to be synchronous
        //
        // from(extensionAPI.workspace.openedTextDocuments)
        //     .pipe(
        //         tap(() => console.log('QQQQQQQQQQQ')),
        //         take(initModel.editors.length)
        //     )
        //     .toPromise(),
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

class AsyncMessagePort extends MessagePort implements MessagePort {
    public get onmessage(): ((this: MessagePort, ev: MessageEvent) => any) | null {
        throw new Error('not implemented')
    }

    public set onmessage(callback: ((this: MessagePort, ev: MessageEvent) => any) | null) {
        // throw new Error('not implemented')
    }

    public get onmessageerror(): ((this: MessagePort, ev: MessageEvent) => any) | null {
        throw new Error('not implemented')
    }

    public set onmessageerror(callback: ((this: MessagePort, ev: MessageEvent) => any) | null) {
        // throw new Error('not implemented')
    }

    private isStarted = false
    private isClosed = false
    private messageListeners: EventListenerOrEventListenerObject[] = []

    public close(): void {
        this.isClosed = true
        this.messageListeners = []
    }

    public start(): void {
        this.isStarted = true
    }

    public postMessage(message: any, transfer?: Transferable[]): void {
        if (!this.isStarted) {
            throw new Error('MessagePort is not started')
        }
        if (this.isClosed) {
            throw new Error('MessagePort is closed')
        }
        this.dispatchEvent(new MessageEvent('message', { data: message }))
    }

    public dispatchEvent(event: Event): boolean {
        if (event.type !== 'message') {
            throw new Error('not implemented')
        }
        for (const listener of this.messageListeners) {
            const handler = 'handleEvent' in listener ? listener.handleEvent.bind(this) : listener
            setTimeout(() => handler(event), 0)
        }
        return true
    }

    public addEventListener<K extends keyof MessagePortEventMap>(
        type: K,
        listener: (this: MessagePort, ev: MessagePortEventMap[K]) => any,
        options?: boolean | AddEventListenerOptions
    ): void
    public addEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | AddEventListenerOptions
    ): void
    public addEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | AddEventListenerOptions
    ): void {
        if (type !== 'message') {
            throw new Error('not implemented')
        }
        this.messageListeners.push(listener)
    }

    public removeEventListener<K extends keyof MessagePortEventMap>(
        type: K,
        listener: (this: MessagePort, ev: MessagePortEventMap[K]) => any,
        options?: boolean | EventListenerOptions
    ): void
    public removeEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | EventListenerOptions
    ): void
    public removeEventListener(
        type: string,
        listener: EventListenerOrEventListenerObject,
        options?: boolean | EventListenerOptions
    ): void {
        if (type !== 'message') {
            throw new Error('not implemented')
        }
        const index = this.messageListeners.indexOf(listener)
        if (index !== -1) {
            this.messageListeners.splice(index, 1)
        }
    }
}

function createMessageChannel(): MessageChannel {
    // return new MessageChannel()
    const port1 = new AsyncMessagePort()
    const port2 = new AsyncMessagePort()
    return { port1, port2 }
}
