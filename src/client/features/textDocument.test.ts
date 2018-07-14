import * as assert from 'assert'
import { BehaviorSubject } from 'rxjs'
import { TextDocument } from 'vscode-languageserver-types'
import { createObservableEnvironment, EMPTY_ENVIRONMENT, Environment } from '../../environment/environment'
import { NotificationType } from '../../jsonrpc2/messages'
import {
    ClientCapabilities,
    DidOpenTextDocumentNotification,
    DidOpenTextDocumentParams,
    TextDocumentRegistrationOptions,
} from '../../protocol'
import { DocumentSelector } from '../../types/document'
import { Client } from '../client'
import { TextDocumentDidOpenFeature } from './textDocument'

const create = (): {
    client: Client
    environment: BehaviorSubject<Environment>
    feature: TextDocumentDidOpenFeature & { readonly selectors: Map<string, DocumentSelector> }
} => {
    const environment = new BehaviorSubject<Environment>(EMPTY_ENVIRONMENT)
    const client = {
        clientOptions: {
            environment: createObservableEnvironment(environment),
            middleware: {},
        },
    } as Client
    const feature = new class extends TextDocumentDidOpenFeature {
        constructor(client: Client) {
            super(client)
        }
        public readonly selectors!: Map<string, DocumentSelector>
    }(client)
    return { client, environment, feature }
}

describe('TextDocumentDidOpenFeature', () => {
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
    })

    describe('when a text document is opened', () => {
        it('sends a textDocument/didOpen notification to the server', done => {
            const { client, environment, feature } = create()
            feature.initialize({ textDocumentSync: { openClose: true } }, ['l'])

            const textDocument = TextDocument.create('file:///f', 'l', 1, 't')

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
                assert.deepStrictEqual(params, {
                    textDocument: {
                        uri: textDocument.uri,
                        languageId: textDocument.languageId,
                        version: textDocument.version,
                        text: textDocument.getText(),
                    },
                } as DidOpenTextDocumentParams)
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
