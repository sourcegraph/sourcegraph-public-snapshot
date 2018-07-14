import { CommandRegistry } from './providers/command'
import { TextDocumentDecorationProviderRegistry } from './providers/decoration'
import { TextDocumentHoverProviderRegistry } from './providers/hover'
import { TextDocumentLocationProviderRegistry } from './providers/location'

/** Registries is a container for all provider registries. */
export class Registries {
    public readonly commands = new CommandRegistry()
    public readonly textDocumentDefinition = new TextDocumentLocationProviderRegistry()
    public readonly textDocumentImplementation = new TextDocumentLocationProviderRegistry()
    public readonly textDocumentTypeDefinition = new TextDocumentLocationProviderRegistry()
    public readonly textDocumentHover = new TextDocumentHoverProviderRegistry()
    public readonly textDocumentDecoration = new TextDocumentDecorationProviderRegistry()
}
