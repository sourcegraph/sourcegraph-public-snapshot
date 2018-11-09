import { Observable } from 'rxjs'
import { SettingsCascade } from '../protocol'
import { Environment } from './environment'
import { Extension } from './extension'
import { CommandRegistry } from './services/command'
import { ContributionRegistry } from './services/contribution'
import { TextDocumentDecorationProviderRegistry } from './services/decoration'
import { TextDocumentHoverProviderRegistry } from './services/hover'
import { TextDocumentLocationProviderRegistry, TextDocumentReferencesProviderRegistry } from './services/location'
import { QueryTransformerRegistry } from './services/queryTransformer'
import { ViewProviderRegistry } from './services/view'

/**
 * Services is a container for all services.
 *
 * @template X extension type
 * @template C settings cascade type
 */
export class Services<X extends Extension, C extends SettingsCascade> {
    constructor(private environment: Observable<Environment<X, C>>) {}

    public readonly commands = new CommandRegistry()
    public readonly contribution = new ContributionRegistry(this.environment)
    public readonly textDocumentDefinition = new TextDocumentLocationProviderRegistry()
    public readonly textDocumentImplementation = new TextDocumentLocationProviderRegistry()
    public readonly textDocumentReferences = new TextDocumentReferencesProviderRegistry()
    public readonly textDocumentTypeDefinition = new TextDocumentLocationProviderRegistry()
    public readonly textDocumentHover = new TextDocumentHoverProviderRegistry()
    public readonly textDocumentDecoration = new TextDocumentDecorationProviderRegistry()
    public readonly queryTransformer = new QueryTransformerRegistry()
    public readonly views = new ViewProviderRegistry()
}
