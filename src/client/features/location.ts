import { from, Observable, Unsubscribable } from 'rxjs'
import { ProvideTextDocumentLocationSignature } from '../../environment/providers/location'
import { FeatureProviderRegistry } from '../../environment/providers/registry'
import {
    ClientCapabilities,
    DefinitionRequest,
    ImplementationRequest,
    ReferenceParams,
    ReferencesRequest,
    TextDocumentPositionParams,
    TextDocumentRegistrationOptions,
    TypeDefinitionRequest,
} from '../../protocol'
import { Location } from '../../protocol/plainTypes'
import { Client } from '../client'
import { ensure, Feature } from './common'

/**
 * Support for requests that retrieve a list of locations (e.g., textDocument/definition,
 * textDocument/implementation, and textDocument/typeDefinition).
 */
export abstract class TextDocumentLocationFeature<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
> extends Feature<TextDocumentRegistrationOptions> {
    constructor(
        client: Client,
        private registry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentLocationSignature<P, L>
        >
    ) {
        super(client)
    }

    /** Override to modify the client capabilities object before sending to report support for this feature. */
    public abstract fillClientCapabilities(capabilities: ClientCapabilities): void

    protected registerProvider(options: TextDocumentRegistrationOptions): Unsubscribable {
        return this.registry.registerProvider(
            options,
            (params: P): Observable<L | L[] | null> => from(this.client.sendRequest(this.messages, params))
        )
    }

    protected validateRegistrationOptions(data: any): TextDocumentRegistrationOptions {
        const options: TextDocumentRegistrationOptions = data
        if (!options.extensionID) {
            throw new Error('extensionID should be non-empty in registration options')
        }
        return options
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
}
