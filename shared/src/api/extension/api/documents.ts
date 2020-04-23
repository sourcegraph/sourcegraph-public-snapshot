import { ProxyMarked, proxyMarker } from '@sourcegraph/comlink'
import { Subject } from 'rxjs'
import { TextDocument } from 'sourcegraph'
import { TextModelUpdate } from '../../client/services/modelService'
import { ExtDocument } from './textDocument'

/** @internal */
export interface ExtDocumentsAPI extends ProxyMarked {
    $acceptDocumentData(modelUpdates: readonly TextModelUpdate[]): void
}

const DOCUMENT_NOT_FOUND_ERROR_NAME = 'DocumentNotFoundError'
class DocumentNotFoundError extends Error {
    public readonly name = DOCUMENT_NOT_FOUND_ERROR_NAME
    constructor(resource: string) {
        super(`document not found: ${resource}`)
    }
}

/** @internal */
export class ExtDocuments implements ExtDocumentsAPI, ProxyMarked {
    public readonly [proxyMarker] = true

    private documents = new Map<string, ExtDocument>()

    constructor(private sync: () => Promise<void>) {}

    /**
     * Returns the known document with the given URI.
     *
     * @internal
     */
    public get(resource: string): ExtDocument {
        const doc = this.documents.get(resource)
        if (!doc) {
            throw new DocumentNotFoundError(resource)
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
    public async getSync(resource: string): Promise<ExtDocument> {
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
    public getAll(): ExtDocument[] {
        return Array.from(this.documents.values())
    }

    public openedTextDocuments = new Subject<TextDocument>()

    public $acceptDocumentData(modelUpdates: readonly TextModelUpdate[]): void {
        for (const update of modelUpdates) {
            switch (update.type) {
                case 'added': {
                    const { uri, languageId, text } = update
                    const doc = new ExtDocument({ uri, languageId, text })
                    this.documents.set(update.uri, doc)
                    this.openedTextDocuments.next(doc)
                    break
                }
                case 'updated': {
                    const doc = this.get(update.uri)
                    doc.update(update)
                    break
                }
                case 'deleted':
                    this.documents.delete(update.uri)
                    break
            }
        }
    }
}
