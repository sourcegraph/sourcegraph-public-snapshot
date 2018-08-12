import * as assert from 'assert'
import { BehaviorSubject, Subject } from 'rxjs'
import { TextDocumentItem } from 'vscode-languageserver-types'
import { createObservableEnvironment, EMPTY_ENVIRONMENT, Environment } from '../../environment/environment'
import { NotificationType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    DidCloseTextDocumentNotification,
    DidCloseTextDocumentParams,
    DidOpenTextDocumentNotification,
    DidOpenTextDocumentParams,
    TextDocumentRegistrationOptions,
    TextDocumentSyncKind,
} from '../../protocol'
import { DocumentSelector } from '../../types/document'
import { Client } from '../client'
import {
    resolveTextDocumentSync,
    TextDocumentDidCloseFeature,
    TextDocumentDidOpenFeature,
    TextDocumentNotificationFeature as AbstractTextDocumentNotificationFeature,
} from './textDocument'

describe('TextDocumentNotificationFeature', () => {
    const create = <F extends AbstractTextDocumentNotificationFeature<any, any>>(
        FeatureClass: new (client: Client) => F
    ): {
        client: Client
        feature: F
    } => {
        const client = {} as Client
        const feature = new FeatureClass(client)
        return { client, feature }
    }

    class TextDocumentNotificationFeature extends AbstractTextDocumentNotificationFeature<any, any> {
        constructor(client: Client) {
            super(client, new Subject<any>(), DidOpenTextDocumentNotification.type, undefined, () => void 0)
        }
        public readonly messages = { method: 'm' }
        public fillClientCapabilities(): void {
            /* noop */
        }
        public initialize(): void {
            /* noop */
        }
    }

    const FIXTURE_REGISTER_OPTIONS: TextDocumentRegistrationOptions = { documentSelector: ['*'] }

    describe('registration', () => {
        it('supports dynamic registration and unregistration', () => {
            const { feature } = create(TextDocumentNotificationFeature)
            feature.register(feature.messages, { id: 'a', registerOptions: FIXTURE_REGISTER_OPTIONS })
            feature.unregister('a')
        })

        it('supports multiple dynamic registrations and unregistrations', () => {
            const { feature } = create(TextDocumentNotificationFeature)
            feature.register(feature.messages, { id: 'a', registerOptions: FIXTURE_REGISTER_OPTIONS })
            feature.register(feature.messages, { id: 'b', registerOptions: FIXTURE_REGISTER_OPTIONS })
            feature.unregister('b')
            feature.unregister('a')
        })

        it('prevents registration with conflicting IDs', () => {
            const { feature } = create(TextDocumentNotificationFeature)
            feature.register(feature.messages, { id: 'a', registerOptions: FIXTURE_REGISTER_OPTIONS })
            assert.throws(() => {
                feature.register(feature.messages, { id: 'a', registerOptions: FIXTURE_REGISTER_OPTIONS })
            })
        })

        it('throws an error if ID to unregister is not registered', () => {
            const { feature } = create(TextDocumentNotificationFeature)
            assert.throws(() => feature.unregister('a'))
        })
    })
})

describe('TextDocumentDidOpenFeature', () => {
    const create = (): {
        client: Client
        environment: BehaviorSubject<Environment>
        feature: TextDocumentDidOpenFeature & { readonly selectors: Map<string, DocumentSelector> }
    } => {
        const environment = new BehaviorSubject<Environment>(EMPTY_ENVIRONMENT)
        const client = { options: { middleware: {} } } as Client
        const feature = new class extends TextDocumentDidOpenFeature {
            public readonly selectors!: Map<string, DocumentSelector>
        }(client, createObservableEnvironment(environment).textDocument)
        return { client, environment, feature }
    }

    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            textDocument: { synchronization: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })

    describe('upon initialization', () => {
        it('registers the provider if the server supports text document sync', () => {
            const { feature } = create()
            feature.initialize({ textDocumentSync: { openClose: true } }, ['*'])
            assert.strictEqual(feature.selectors.size, 1)
        })

        it('does not register the provider if the server lacks support for text document sync', () => {
            const { feature } = create()
            feature.initialize({ textDocumentSync: { openClose: false } }, ['*'])
            assert.strictEqual(feature.selectors.size, 0)
        })

        it('does not register the provider if the server omits mention of support for text document sync', () => {
            const { feature } = create()
            feature.initialize({}, ['*'])
            assert.strictEqual(feature.selectors.size, 0)
        })

        it('registers the provider if the server supports text document sync (backcompat for non-TextDocumentSyncOptions value)', () => {
            const { feature } = create()
            feature.initialize({ textDocumentSync: TextDocumentSyncKind.Incremental }, ['*'])
            assert.strictEqual(feature.selectors.size, 1)
        })

        it('does not register the provider if the server omits mention of support for text document sync (backcompat for non-TextDocumentSyncOptions value)', () => {
            const { feature } = create()
            feature.initialize({ textDocumentSync: TextDocumentSyncKind.None }, ['*'])
            assert.strictEqual(feature.selectors.size, 0)
        })
    })

    describe('when a text document is opened', () => {
        it('sends a textDocument/didOpen notification to the server', done => {
            const { client, environment, feature } = create()
            feature.initialize({ textDocumentSync: { openClose: true } }, ['l'])

            const textDocument: TextDocumentItem = {
                uri: 'file:///f',
                languageId: 'l',
                version: 1,
                text: '',
            }

            function mockSendNotification(method: string, params: any): void
            function mockSendNotification(
                type: NotificationType<DidOpenTextDocumentParams, TextDocumentRegistrationOptions>,
                params: DidOpenTextDocumentParams
            ): void
            function mockSendNotification(
                type: string | NotificationType<DidOpenTextDocumentParams, TextDocumentRegistrationOptions>,
                params: any
            ): void {
                assert.strictEqual(
                    typeof type === 'string' ? type : type.method,
                    DidOpenTextDocumentNotification.type.method
                )
                assert.deepStrictEqual(params, { textDocument } as DidOpenTextDocumentParams)
                done()
            }
            client.sendNotification = mockSendNotification

            environment.next({
                ...environment.value,
                component: { document: textDocument, selections: [], visibleRanges: [] },
            })
        })
    })
})

describe('TextDocumentDidCloseFeature', () => {
    const create = (): {
        client: Client
        environment: BehaviorSubject<Environment>
        feature: TextDocumentDidCloseFeature & { readonly selectors: Map<string, DocumentSelector> }
    } => {
        const environment = new BehaviorSubject<Environment>(EMPTY_ENVIRONMENT)
        const client = { options: { middleware: {} } } as Client
        const feature = new class extends TextDocumentDidCloseFeature {
            public readonly selectors!: Map<string, DocumentSelector>
        }(client, createObservableEnvironment(environment).textDocument)
        return { client, environment, feature }
    }

    it('reports client capabilities', () => {
        const capabilities: ClientCapabilities = {}
        create().feature.fillClientCapabilities(capabilities)
        assert.deepStrictEqual(capabilities, {
            textDocument: { synchronization: { dynamicRegistration: true } },
        } as ClientCapabilities)
    })

    describe('upon initialization', () => {
        it('registers the provider if the server supports text document sync', () => {
            const { feature } = create()
            feature.initialize({ textDocumentSync: { openClose: true } }, ['*'])
            assert.strictEqual(feature.selectors.size, 1)
        })

        it('does not register the provider if the server lacks support for text document sync', () => {
            const { feature } = create()
            feature.initialize({ textDocumentSync: { openClose: false } }, ['*'])
            assert.strictEqual(feature.selectors.size, 0)
        })

        it('does not register the provider if the server omits mention of support for text document sync', () => {
            const { feature } = create()
            feature.initialize({}, ['*'])
            assert.strictEqual(feature.selectors.size, 0)
        })

        it('registers the provider if the server supports text document sync (backcompat for non-TextDocumentSyncOptions value)', () => {
            const { feature } = create()
            feature.initialize({ textDocumentSync: TextDocumentSyncKind.Incremental }, ['*'])
            assert.strictEqual(feature.selectors.size, 1)
        })

        it('does not register the provider if the server omits mention of support for text document sync (backcompat for non-TextDocumentSyncOptions value)', () => {
            const { feature } = create()
            feature.initialize({ textDocumentSync: TextDocumentSyncKind.None }, ['*'])
            assert.strictEqual(feature.selectors.size, 0)
        })
    })

    describe('when a text document is opened and then closed', () => {
        it('sends a textDocument/didClose notification to the server', done => {
            const { client, environment, feature } = create()
            feature.initialize({ textDocumentSync: { openClose: true } }, ['l'])

            const textDocument: TextDocumentItem = {
                uri: 'file:///f',
                languageId: 'l',
                version: 1,
                text: '',
            }

            let didCloseNotifications: DidCloseTextDocumentParams[] = []
            function mockSendNotification(method: string, params: any): void
            function mockSendNotification(
                type: NotificationType<DidCloseTextDocumentParams, TextDocumentRegistrationOptions>,
                params: DidCloseTextDocumentParams
            ): void
            function mockSendNotification(
                type: string | NotificationType<DidCloseTextDocumentParams, TextDocumentRegistrationOptions>,
                params: any
            ): void {
                assert.strictEqual(
                    typeof type === 'string' ? type : type.method,
                    DidCloseTextDocumentNotification.type.method
                )
                didCloseNotifications.push(params)
            }
            client.sendNotification = mockSendNotification

            // Open the document.
            environment.next({
                ...environment.value,
                component: { document: textDocument, selections: [], visibleRanges: [] },
            })
            assert.deepStrictEqual(didCloseNotifications, [])
            didCloseNotifications = []

            // Close the document by setting component to null.
            environment.next({
                ...environment.value,
                component: null,
            })
            assert.deepStrictEqual(didCloseNotifications, [
                {
                    textDocument: { uri: textDocument.uri },
                },
            ] as DidCloseTextDocumentParams[])
            didCloseNotifications = []

            // Reopen the document.
            environment.next({
                ...environment.value,
                component: { document: textDocument, selections: [], visibleRanges: [] },
            })
            assert.deepStrictEqual(didCloseNotifications, [])
            didCloseNotifications = []

            // Close the document by setting component to a different document.
            environment.next({
                ...environment.value,
                component: { document: { ...textDocument, uri: 'file:///f2' }, selections: [], visibleRanges: [] },
            })
            assert.deepStrictEqual(didCloseNotifications, [
                {
                    textDocument: { uri: textDocument.uri },
                },
            ] as DidCloseTextDocumentParams[])
            didCloseNotifications = []

            done()
        })
    })
})

describe('resolveTextDocumentSync', () => {
    it('resolves to TextDocumentSyncOptions', () => {
        assert.deepStrictEqual(resolveTextDocumentSync(undefined), undefined)
        assert.deepStrictEqual(resolveTextDocumentSync(null as any), undefined)
        assert.deepStrictEqual(resolveTextDocumentSync(TextDocumentSyncKind.None), undefined)
        assert.deepStrictEqual(resolveTextDocumentSync(TextDocumentSyncKind.Incremental), {
            openClose: true,
            change: TextDocumentSyncKind.Incremental,
            save: { includeText: false },
        })
        assert.deepStrictEqual(resolveTextDocumentSync(TextDocumentSyncKind.Full), {
            openClose: true,
            change: TextDocumentSyncKind.Full,
            save: { includeText: false },
        })
        assert.deepStrictEqual({ openClose: true }, { openClose: true })
    })
})
