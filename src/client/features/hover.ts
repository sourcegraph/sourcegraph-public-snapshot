import { from, Observable, Unsubscribable } from 'rxjs'
import uuidv4 from 'uuid/v4'
import { Hover, MarkupKind } from 'vscode-languageserver-types'
import { ProvideTextDocumentHoverSignature } from '../../environment/providers/hover'
import { FeatureProviderRegistry } from '../../environment/providers/registry'
import {
    ClientCapabilities,
    HoverRequest,
    ServerCapabilities,
    TextDocumentPositionParams,
    TextDocumentRegistrationOptions,
} from '../../protocol'
import { DocumentSelector } from '../../types/document'
import { Client } from '../client'
import { ensure, Feature } from './common'

/**
 * Support for hover messages (textDocument/hover requests to the server).
 */
export class TextDocumentHoverFeature extends Feature<TextDocumentRegistrationOptions> {
    constructor(
        client: Client,
        private registry: FeatureProviderRegistry<TextDocumentRegistrationOptions, ProvideTextDocumentHoverSignature>
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
        return this.registry.registerProvider(
            options,
            (params: TextDocumentPositionParams): Observable<Hover | null> =>
                from(this.client.sendRequest(HoverRequest.type, params))
        )
    }
}
