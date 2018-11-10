import { Observable } from 'rxjs';
import { ConfigurationCascade } from '../protocol';
import { Environment } from './environment';
import { Extension } from './extension';
import { CommandRegistry } from './providers/command';
import { ContributionRegistry } from './providers/contribution';
import { TextDocumentDecorationProviderRegistry } from './providers/decoration';
import { TextDocumentHoverProviderRegistry } from './providers/hover';
import { TextDocumentLocationProviderRegistry, TextDocumentReferencesProviderRegistry } from './providers/location';
import { QueryTransformerRegistry } from './providers/queryTransformer';
import { ViewProviderRegistry } from './providers/view';
/**
 * Registries is a container for all provider registries.
 *
 * @template X extension type
 * @template C configuration cascade type
 */
export declare class Registries<X extends Extension, C extends ConfigurationCascade> {
    private environment;
    constructor(environment: Observable<Environment<X, C>>);
    readonly commands: CommandRegistry;
    readonly contribution: ContributionRegistry;
    readonly textDocumentDefinition: TextDocumentLocationProviderRegistry<import("../protocol/textDocument").TextDocumentPositionParams, import("../protocol/plainTypes").Location>;
    readonly textDocumentImplementation: TextDocumentLocationProviderRegistry<import("../protocol/textDocument").TextDocumentPositionParams, import("../protocol/plainTypes").Location>;
    readonly textDocumentReferences: TextDocumentReferencesProviderRegistry;
    readonly textDocumentTypeDefinition: TextDocumentLocationProviderRegistry<import("../protocol/textDocument").TextDocumentPositionParams, import("../protocol/plainTypes").Location>;
    readonly textDocumentHover: TextDocumentHoverProviderRegistry;
    readonly textDocumentDecoration: TextDocumentDecorationProviderRegistry;
    readonly queryTransformer: QueryTransformerRegistry;
    readonly views: ViewProviderRegistry;
}
