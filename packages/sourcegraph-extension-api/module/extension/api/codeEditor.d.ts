import * as sourcegraph from 'sourcegraph';
import { ClientCodeEditorAPI } from '../../client/api/codeEditor';
import { ExtDocuments } from './documents';
/** @internal */
export declare class ExtCodeEditor implements sourcegraph.CodeEditor {
    private resource;
    private proxy;
    private documents;
    constructor(resource: string, proxy: ClientCodeEditorAPI, documents: ExtDocuments);
    readonly type = "CodeEditor";
    readonly document: sourcegraph.TextDocument;
    setDecorations(_decorationType: null, decorations: sourcegraph.TextDocumentDecoration[]): void;
    toJSON(): any;
}
