import { DocumentSelector } from 'sourcegraph';
/**
 * A literal to identify a text document in the client.
 */
export interface TextDocumentIdentifier {
    /**
     * The text document's uri.
     */
    uri: string;
}
/**
 * An item to transfer a text document from the client to the server.
 */
export interface TextDocumentItem {
    uri: string;
    languageId: string;
    text: string;
}
export declare function match(selectors: DocumentSelector | IterableIterator<DocumentSelector>, document: TextDocumentItem): boolean;
/**
 * Taken from
 * https://github.com/Microsoft/vscode/blob/3d35801127f0a62d58d752bc613506e836c5d120/src/vs/editor/common/modes/languageSelector.ts#L24.
 */
export declare function score(selector: DocumentSelector, candidateUri: string, candidateLanguage: string): number;
