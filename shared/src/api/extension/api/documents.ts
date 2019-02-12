import { ProxyValue, proxyValueSymbol } from 'comlink'
import { Observable, Subject } from 'rxjs'
import { TextDocument } from 'sourcegraph'
import { TextDocumentItem } from '../../client/types/textDocument'

/** @internal */
export interface ExtDocumentsAPI {
    $acceptDocumentData(doc: TextDocumentItem[]): void
}

/** @internal */
export class ExtDocuments implements ExtDocumentsAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private documents = new Map<string, TextDocumentItem>()

    constructor(private sync: () => Promise<void>) {}

    /**
     * Returns the known document with the given URI.
     *
     * @internal
     */
    public get(resource: string): TextDocument {
        const doc = this.documents.get(resource)
        if (!doc) {
            throw new Error(`document not found: ${resource}`)
        }
        return doc
    }

    /**
     * If needed, perform a sync with the client to ensure that its pending sends have been received before
     * retrieving this document.
     *
     * @todo This is necessary because hovers can be sent before the document is loaded, and it will cause a
     * "document not found" error.
     */
    public async getSync(resource: string): Promise<TextDocument> {
        const doc = this.documents.get(resource)
        if (doc) {
            return doc
        }
        await this.sync()
        return this.get(resource)
    }

    /**
     * Returns all known documents.
     *
     * @internal
     */
    public getAll(): TextDocument[] {
        return Array.from(this.documents.values())
    }

    private textDocumentAdds = new Subject<TextDocument>()
    public readonly onDidOpenTextDocument: Observable<TextDocument> = this.textDocumentAdds

    public $acceptDocumentData(docs: TextDocumentItem[] | null): void {
        if (!docs) {
            // We don't ever (yet) communicate to the extension when docs are closed.
            return
        }
        for (const doc of docs) {
            const isNew = !this.documents.has(doc.uri)
            this.documents.set(doc.uri, doc)
            if (isNew) {
                this.textDocumentAdds.next(doc)
            }
        }
    }
}
