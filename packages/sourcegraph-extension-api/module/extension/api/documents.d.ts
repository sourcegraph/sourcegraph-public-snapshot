import { Observable } from 'rxjs';
import { TextDocument } from 'sourcegraph';
import { TextDocumentItem } from '../../client/types/textDocument';
/** @internal */
export interface ExtDocumentsAPI {
    $acceptDocumentData(doc: TextDocumentItem[]): void;
}
/** @internal */
export declare class ExtDocuments implements ExtDocumentsAPI {
    private sync;
    private documents;
    constructor(sync: () => Promise<void>);
    /**
     * Returns the known document with the given URI.
     *
     * @internal
     */
    get(resource: string): TextDocument;
    /**
     * If needed, perform a sync with the client to ensure that its pending sends have been received before
     * retrieving this document.
     *
     * @todo This is necessary because hovers can be sent before the document is loaded, and it will cause a
     * "document not found" error.
     */
    getSync(resource: string): Promise<TextDocument>;
    /**
     * Returns all known documents.
     *
     * @internal
     */
    getAll(): TextDocument[];
    private textDocumentAdds;
    readonly onDidOpenTextDocument: Observable<TextDocument>;
    $acceptDocumentData(docs: TextDocumentItem[] | null): void;
}
