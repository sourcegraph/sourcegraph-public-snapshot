import { Unsubscribable } from 'rxjs';
import { DefinitionProvider, DocumentSelector, HoverProvider, ImplementationProvider, ReferenceContext, ReferenceProvider, TypeDefinitionProvider } from 'sourcegraph';
import { ClientLanguageFeaturesAPI } from '../../client/api/languageFeatures';
import * as plain from '../../protocol/plainTypes';
import { ExtDocuments } from './documents';
/** @internal */
export interface ExtLanguageFeaturesAPI {
    $provideHover(id: number, resource: string, position: plain.Position): Promise<plain.Hover | null | undefined>;
    $provideDefinition(id: number, resource: string, position: plain.Position): Promise<plain.Definition | undefined>;
    $provideTypeDefinition(id: number, resource: string, position: plain.Position): Promise<plain.Definition | undefined>;
    $provideImplementation(id: number, resource: string, position: plain.Position): Promise<plain.Definition | undefined>;
    $provideReferences(id: number, resource: string, position: plain.Position, context: ReferenceContext): Promise<plain.Location[] | null | undefined>;
}
/** @internal */
export declare class ExtLanguageFeatures implements ExtLanguageFeaturesAPI {
    private proxy;
    private documents;
    private registrations;
    constructor(proxy: ClientLanguageFeaturesAPI, documents: ExtDocuments);
    $provideHover(id: number, resource: string, position: plain.Position): Promise<plain.Hover | null | undefined>;
    registerHoverProvider(selector: DocumentSelector, provider: HoverProvider): Unsubscribable;
    $provideDefinition(id: number, resource: string, position: plain.Position): Promise<plain.Definition | null | undefined>;
    registerDefinitionProvider(selector: DocumentSelector, provider: DefinitionProvider): Unsubscribable;
    $provideTypeDefinition(id: number, resource: string, position: plain.Position): Promise<plain.Definition | null | undefined>;
    registerTypeDefinitionProvider(selector: DocumentSelector, provider: TypeDefinitionProvider): Unsubscribable;
    $provideImplementation(id: number, resource: string, position: plain.Position): Promise<plain.Definition | undefined>;
    registerImplementationProvider(selector: DocumentSelector, provider: ImplementationProvider): Unsubscribable;
    $provideReferences(id: number, resource: string, position: plain.Position, context: ReferenceContext): Promise<plain.Location[] | null | undefined>;
    registerReferenceProvider(selector: DocumentSelector, provider: ReferenceProvider): Unsubscribable;
}
