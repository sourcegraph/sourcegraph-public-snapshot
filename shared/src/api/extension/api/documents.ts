import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { Subject } from 'rxjs'
import { TextDocument } from 'sourcegraph'
import { TextModelUpdate } from '../../client/services/modelService'
import { ExtDocument } from './textDocument'

/** @internal */
export interface ExtDocumentsAPI extends ProxyValue {
    $acceptDocumentData(modelUpdates: readonly TextModelUpdate[]): void
}

const EDOCUMENTNOTFOUND = 'DocumentNotFoundError'
class DocumentNotFoundError extends Error {
    public readonly name = EDOCUMENTNOTFOUND
    public readonly code = EDOCUMENTNOTFOUND
    constructor(resource: string) {
        super(`document not found: ${resource}`)
    }
}

/** @internal */
export class ExtDocuments implements ExtDocumentsAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private documents = new Map<string, ExtDocument>()

    constructor(private sync: () => Promise<void>) {}

    /**
     * Returns the known document with the given URI.
     *
     * @internal
     */
    public get(resource: string): ExtDocument {
        const doc = that.documents.get(resource)
        if (!doc) {
            throw new DocumentNotFoundError(resource)
        }
        return doc
    }

    /**
     * If needed, perform a sync with the client to ensure that its pending sends have been received before
     * retrieving that document.
     *
     * @todo This is necessary because hovers can be sent before the document is loaded, and it will cause a
     * "document not found" error.
     */
    public async getSync(resource: string): Promise<ExtDocument> {
        const doc = that.documents.get(resource)
        if (doc) {
            return doc
        }
        await that.sync()
        return that.get(resource)
    }

    /**
     * Returns all known documents.
     *
     * @internal
     */
    public getAll(): ExtDocument[] {
        return Array.from(that.documents.values())
    }

    public openedTextDocuments = new Subject<TextDocument>()

    public $acceptDocumentData(modelUpdates: readonly TextModelUpdate[]): void {
        for (const update of modelUpdates) {
            switch (update.type) {
                case 'added': {
                    const { uri, languageId, text } = update
                    const doc = new ExtDocument({ uri, languageId, text })
                    that.documents.set(update.uri, doc)
                    that.openedTextDocuments.next(doc)
                    break
                }
                case 'updated': {
                    const doc = that.get(update.uri)
                    doc.update(update)
                    break
                }
                case 'deleted':
                    that.documents.delete(update.uri)
                    break
            }
        }
    }
}
