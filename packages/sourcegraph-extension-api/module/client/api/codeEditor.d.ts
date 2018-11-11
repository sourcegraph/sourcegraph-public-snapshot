import { Connection } from '../../protocol/jsonrpc2/connection';
import { TextDocumentDecoration } from '../../protocol/plainTypes';
import { ProvideTextDocumentDecorationSignature } from '../providers/decoration';
import { FeatureProviderRegistry } from '../providers/registry';
/** @internal */
export interface ClientCodeEditorAPI {
    $setDecorations(resource: string, decorations: TextDocumentDecoration[]): void;
}
/** @internal */
export declare class ClientCodeEditor implements ClientCodeEditorAPI {
    private registry;
    private subscriptions;
    /** Map of document URI to its decorations (last published by the server). */
    private decorations;
    constructor(connection: Connection, registry: FeatureProviderRegistry<undefined, ProvideTextDocumentDecorationSignature>);
    $setDecorations(resource: string, decorations: TextDocumentDecoration[]): void;
    private getDecorationsSubject;
    unsubscribe(): void;
}
