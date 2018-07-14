import { CommandRegistry } from './providers/command'
import { TextDocumentDecorationProviderRegistry } from './providers/decoration'
import { TextDocumentDefinitionProviderRegistry } from './providers/definition'
import { TextDocumentHoverProviderRegistry } from './providers/hover'

/** Registries is a container for all provider registries. */
export class Registries {
    public readonly commands = new CommandRegistry()
    public readonly textDocumentDefinition = new TextDocumentDefinitionProviderRegistry()
    public readonly textDocumentHover = new TextDocumentHoverProviderRegistry()
    public readonly textDocumentDecoration = new TextDocumentDecorationProviderRegistry()
}
