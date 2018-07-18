import { Unsubscribable } from 'rxjs'
import uuidv4 from 'uuid/v4'
import { Location } from 'vscode-languageserver-types'
import { ProvideTextDocumentLocationSignature } from '../../environment/providers/location'
import { TextDocumentFeatureProviderRegistry } from '../../environment/providers/textDocument'
import {
    ClientCapabilities,
    DefinitionRequest,
    ImplementationRequest,
    ReferenceParams,
    ReferencesRequest,
    ServerCapabilities,
    TextDocumentPositionParams,
    TextDocumentRegistrationOptions,
    TypeDefinitionRequest,
} from '../../protocol'
import { DocumentSelector } from '../../types/document'
import { NextSignature } from '../../types/middleware'
import { Client } from '../client'
import { Middleware } from '../middleware'
import { ensure, TextDocumentFeature } from './common'

export type ProvideTextDocumentLocationMiddleware<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
> = NextSignature<P, Promise<L | L[] | null>>

/**
 * Support for requests that retrieve a list of locations (e.g., textDocument/definition,
 * textDocument/implementation, and textDocument/typeDefinition).
 */
export abstract class TextDocumentLocationFeature<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
> extends TextDocumentFeature<TextDocumentRegistrationOptions> {
    constructor(
        client: Client,
        private registry: TextDocumentFeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentLocationSignature<P, L>
        >
    ) {
        super(client)
    }

    /** Override to modify the client capabilities object before sending to report support for this feature. */
    public abstract fillClientCapabilities(capabilities: ClientCapabilities): void

    /** Override to compute whether the server capabilities report support for this feature. */
    protected abstract isSupported(capabilities: ServerCapabilities): boolean

    /** Override to return the middleware for this feature. */
    protected abstract getMiddleware?(midleware: Middleware): ProvideTextDocumentLocationMiddleware<P, L> | undefined

    public initialize(capabilities: ServerCapabilities, documentSelector: DocumentSelector): void {
        if (!this.isSupported(capabilities) || !documentSelector) {
            return
        }
        this.register(this.messages, {
            id: uuidv4(),
            registerOptions: { documentSelector },
        })
    }

    protected registerProvider(options: TextDocumentRegistrationOptions): Unsubscribable {
        const client = this.client
        const provideTextDocumentLocation: ProvideTextDocumentLocationSignature<P, L> = params =>
            client.sendRequest(this.messages, params)
        const middleware = this.getMiddleware ? this.getMiddleware(client.options.middleware) : undefined
        return this.registry.registerProvider(
            options,
            (params: P): Promise<L | L[] | null> =>
                middleware ? middleware(params, provideTextDocumentLocation) : provideTextDocumentLocation(params)
        )
    }
}

/**
 * Support for definition requests (textDocument/definition requests to the server).
 */
export class TextDocumentDefinitionFeature extends TextDocumentLocationFeature {
    public readonly messages = DefinitionRequest.type

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        const capability = ensure(ensure(capabilities, 'textDocument')!, 'definition')!
        capability.dynamicRegistration = true
    }

    protected isSupported(capabilities: ServerCapabilities): boolean {
        return !!capabilities.definitionProvider
    }

    protected getMiddleware(
        middleware: Middleware
    ): ProvideTextDocumentLocationMiddleware<TextDocumentPositionParams, Location> | undefined {
        return middleware.provideTextDocumentDefinition
    }
}

/**
 * Support for implementation requests (textDocument/implementation requests to the server).
 */
export class TextDocumentImplementationFeature extends TextDocumentLocationFeature {
    public readonly messages = ImplementationRequest.type

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        const capability = ensure(ensure(capabilities, 'textDocument')!, 'implementation')!
        capability.dynamicRegistration = true
    }

    protected isSupported(capabilities: ServerCapabilities): boolean {
        return !!capabilities.implementationProvider
    }

    protected getMiddleware(
        middleware: Middleware
    ): ProvideTextDocumentLocationMiddleware<TextDocumentPositionParams, Location> | undefined {
        return middleware.provideTextDocumentImplementation
    }
}

/**
 * Support for type definition requests (textDocument/typeDefinition requests to the server).
 */
export class TextDocumentTypeDefinitionFeature extends TextDocumentLocationFeature {
    public readonly messages = TypeDefinitionRequest.type

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        const capability = ensure(ensure(capabilities, 'textDocument')!, 'typeDefinition')!
        capability.dynamicRegistration = true
    }

    protected isSupported(capabilities: ServerCapabilities): boolean {
        return !!capabilities.typeDefinitionProvider
    }

    protected getMiddleware(
        middleware: Middleware
    ): ProvideTextDocumentLocationMiddleware<TextDocumentPositionParams, Location> | undefined {
        return middleware.provideTextDocumentTypeDefinition
    }
}

/**
 * Support for references requests (textDocument/references requests to the server).
 */
export class TextDocumentReferencesFeature extends TextDocumentLocationFeature<ReferenceParams, Location> {
    public readonly messages = ReferencesRequest.type

    public fillClientCapabilities(capabilities: ClientCapabilities): void {
        const capability = ensure(ensure(capabilities, 'textDocument')!, 'references')!
        capability.dynamicRegistration = true
    }

    protected isSupported(capabilities: ServerCapabilities): boolean {
        return !!capabilities.referencesProvider
    }

    protected getMiddleware(
        middleware: Middleware
    ): ProvideTextDocumentLocationMiddleware<ReferenceParams> | undefined {
        return middleware.provideTextDocumentReferences
    }
}
