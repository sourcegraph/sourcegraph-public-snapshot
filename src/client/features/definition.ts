import { TeardownLogic } from 'rxjs'
import * as uuidv4 from 'uuid/v4'
import { Definition } from 'vscode-languageserver-types'
import { ProvideTextDocumentLocationSignature } from '../../environment/providers/location'
import { TextDocumentFeatureProviderRegistry } from '../../environment/providers/textDocument'
import {
    ClientCapabilities,
    DefinitionRequest,
    ServerCapabilities,
    TextDocumentPositionParams,
    TextDocumentRegistrationOptions,
} from '../../protocol'
import { DocumentSelector } from '../../types/document'
import { NextSignature } from '../../types/middleware'
import { Client } from '../client'
import { ensure, TextDocumentFeature } from './common'

export type ProvideTextDocumentDefinitionMiddleware = NextSignature<
    TextDocumentPositionParams,
    Promise<Definition | null>
>

/**
 * Support for go-to-definition (textDocument/definition requests to the server).
 */
export class TextDocumentDefinitionFeature extends TextDocumentFeature<TextDocumentRegistrationOptions> {
    constructor(
        client: Client,
        private registry: TextDocumentFeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentLocationSignature
        >
    ) {
        super(client, DefinitionRequest.type)
    }

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        const definitionCapability = ensure(ensure(capabilities, 'textDocument')!, 'definition')!
        definitionCapability.dynamicRegistration = true
    }

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        if (!capabilities.definitionProvider || !documentSelector) {
            return
        }
        this.register(this.messages, {
            id: uuidv4(),
            registerOptions: { documentSelector },
        })
    }

    protected registerProvider(options: TextDocumentRegistrationOptions): TeardownLogic {
        const client = this.client
        const provideTextDocumentDefinition: ProvideTextDocumentLocationSignature = params =>
            client.sendRequest(DefinitionRequest.type, params)
        const middleware = client.clientOptions.middleware!
        return this.registry.registerProvider(
            options,
            (params: TextDocumentPositionParams): Promise<Definition | null> =>
                middleware.provideTextDocumentDefinition
                    ? middleware.provideTextDocumentDefinition(params, provideTextDocumentDefinition)
                    : provideTextDocumentDefinition(params)
        )
    }
}
