import { Unsubscribable } from 'rxjs'
import uuidv4 from 'uuid/v4'
import { Hover, MarkupKind } from 'vscode-languageserver-types'
import { ProvideTextDocumentHoverSignature } from '../../environment/providers/hover'
import { TextDocumentFeatureProviderRegistry } from '../../environment/providers/textDocument'
import {
    ClientCapabilities,
    HoverRequest,
    ServerCapabilities,
    TextDocumentPositionParams,
    TextDocumentRegistrationOptions,
} from '../../protocol'
import { DocumentSelector } from '../../types/document'
import { NextSignature } from '../../types/middleware'
import { Client } from '../client'
import { ensure, TextDocumentFeature } from './common'

export type ProvideTextDocumentHoverMiddleware = NextSignature<TextDocumentPositionParams, Promise<Hover | null>>

/**
 * Support for hover messages (textDocument/hover requests to the server).
 */
export class TextDocumentHoverFeature extends TextDocumentFeature<TextDocumentRegistrationOptions> {
    constructor(
        client: Client,
        private registry: TextDocumentFeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentHoverSignature
        >
    ) {
        super(client)
    }

    public readonly messages = HoverRequest.type

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        const hoverCapability = ensure(ensure(capabilities, 'textDocument')!, 'hover')!
        hoverCapability.dynamicRegistration = true
        hoverCapability.contentFormat = [MarkupKind.Markdown, MarkupKind.PlainText]
    }

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        if (!capabilities.hoverProvider || !documentSelector) {
            return
        }
        this.register(this.messages, {
            id: uuidv4(),
            registerOptions: { documentSelector },
        })
    }

    protected registerProvider(options: TextDocumentRegistrationOptions): Unsubscribable {
        const client = this.client
        const provideTextDocumentHover: ProvideTextDocumentHoverSignature = params =>
            client.sendRequest(HoverRequest.type, params)
        const middleware = client.options.middleware.provideTextDocumentHover
        return this.registry.registerProvider(
            options,
            (params: TextDocumentPositionParams): Promise<Hover | null> =>
                middleware ? middleware(params, provideTextDocumentHover) : provideTextDocumentHover(params)
        )
    }
}
