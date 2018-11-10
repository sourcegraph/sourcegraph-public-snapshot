import { DocumentSelector } from 'sourcegraph';
import { TextDocumentRegistrationOptions } from '../../protocol';
import { Connection } from '../../protocol/jsonrpc2/connection';
import { ProvideTextDocumentHoverSignature } from '../providers/hover';
import { ProvideTextDocumentLocationSignature, TextDocumentReferencesProviderRegistry } from '../providers/location';
import { FeatureProviderRegistry } from '../providers/registry';
/** @internal */
export interface ClientLanguageFeaturesAPI {
    $unregister(id: number): void;
    $registerHoverProvider(id: number, selector: DocumentSelector): void;
    $registerDefinitionProvider(id: number, selector: DocumentSelector): void;
    $registerTypeDefinitionProvider(id: number, selector: DocumentSelector): void;
    $registerImplementationProvider(id: number, selector: DocumentSelector): void;
    $registerReferenceProvider(id: number, selector: DocumentSelector): void;
}
/** @internal */
export declare class ClientLanguageFeatures implements ClientLanguageFeaturesAPI {
    private hoverRegistry;
    private definitionRegistry;
    private typeDefinitionRegistry;
    private implementationRegistry;
    private referencesRegistry;
    private subscriptions;
    private registrations;
    private proxy;
    constructor(connection: Connection, hoverRegistry: FeatureProviderRegistry<TextDocumentRegistrationOptions, ProvideTextDocumentHoverSignature>, definitionRegistry: FeatureProviderRegistry<TextDocumentRegistrationOptions, ProvideTextDocumentLocationSignature>, typeDefinitionRegistry: FeatureProviderRegistry<TextDocumentRegistrationOptions, ProvideTextDocumentLocationSignature>, implementationRegistry: FeatureProviderRegistry<TextDocumentRegistrationOptions, ProvideTextDocumentLocationSignature>, referencesRegistry: TextDocumentReferencesProviderRegistry);
    $unregister(id: number): void;
    $registerHoverProvider(id: number, selector: DocumentSelector): void;
    $registerDefinitionProvider(id: number, selector: DocumentSelector): void;
    $registerTypeDefinitionProvider(id: number, selector: DocumentSelector): void;
    $registerImplementationProvider(id: number, selector: DocumentSelector): void;
    $registerReferenceProvider(id: number, selector: DocumentSelector): void;
    unsubscribe(): void;
}
