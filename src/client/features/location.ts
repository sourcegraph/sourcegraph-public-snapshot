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
import { Middleware } from '../middleware'
import { ensure, TextDocumentFeature } from './common'

export type ProvideTextDocumentLocationMiddleware = NextSignature<
    TextDocumentPositionParams,
    Promise<Definition | null>
>

/**
 * Support for requests that retrieve a list of locations (e.g., textDocument/definition,
 * textDocument/implementation, and textDocument/typeDefinition).
 */
export abstract class TextDocumentLocationFeature extends TextDocumentFeature<TextDocumentRegistrationOptions> {
    constructor(
        client: Client,
        private registry: TextDocumentFeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentLocationSignature
        >
    ) {
        super(client, DefinitionRequest.type)
    }

    /** Override to modify the client capabilities object before sending to report support for this feature. */
    public abstract fillClientCapabilities(capabilities: ClientCapabilities): void

    /** Override to compute whether the server capabilities report support for this feature. */
    protected abstract isSupported(capabilities: ServerCapabilities): boolean

    /** Override to return the middleware for this feature. */
    protected abstract getMiddleware?(midleware: Middleware): ProvideTextDocumentLocationMiddleware | undefined

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        if (!this.isSupported(capabilities) || !documentSelector) {
            return
        }
        this.register(this.messages, {
            id: uuidv4(),
            registerOptions: { documentSelector },
        })
    }

    protected registerProvider(options: TextDocumentRegistrationOptions): TeardownLogic {
        const client = this.client
        const provideTextDocumentLocation: ProvideTextDocumentLocationSignature = params =>
            client.sendRequest(this.messages, params)
        const middleware = this.getMiddleware ? this.getMiddleware(client.clientOptions.middleware!) : undefined
        return this.registry.registerProvider(
            options,
            (params: TextDocumentPositionParams): Promise<Definition | null> =>
                middleware ? middleware(params, provideTextDocumentLocation) : provideTextDocumentLocation(params)
        )
    }
}

/**
 * Support for definition requests (textDocument/definition requests to the server).
 */
export class TextDocumentDefinitionFeature extends TextDocumentLocationFeature {
    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        const capability = ensure(ensure(capabilities, 'textDocument')!, 'definition')!
        capability.dynamicRegistration = true
    }

    protected isSupported(capabilities: ServerCapabilities): boolean {
        return !!capabilities.definitionProvider
    }

    protected getMiddleware(middleware: Middleware): ProvideTextDocumentLocationMiddleware | undefined {
        return middleware.provideTextDocumentDefinition
    }
}

/**
 * Support for implementation requests (textDocument/implementation requests to the server).
 */
export class TextDocumentImplementationFeature extends TextDocumentLocationFeature {
    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        const capability = ensure(ensure(capabilities, 'textDocument')!, 'implementation')!
        capability.dynamicRegistration = true
    }

    protected isSupported(capabilities: ServerCapabilities): boolean {
        return !!capabilities.implementationProvider
    }

    protected getMiddleware(middleware: Middleware): ProvideTextDocumentLocationMiddleware | undefined {
        return middleware.provideTextDocumentImplementation
    }
}

/**
 * Support for type definition requests (textDocument/typeDefinition requests to the server).
 */
export class TextDocumentTypeDefinitionFeature extends TextDocumentLocationFeature {
    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        const capability = ensure(ensure(capabilities, 'textDocument')!, 'typeDefinition')!
        capability.dynamicRegistration = true
    }

    protected isSupported(capabilities: ServerCapabilities): boolean {
        return !!capabilities.typeDefinitionProvider
    }

    protected getMiddleware(middleware: Middleware): ProvideTextDocumentLocationMiddleware | undefined {
        return middleware.provideTextDocumentTypeDefinition
    }
}
