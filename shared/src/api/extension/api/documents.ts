import { proxyMarker, ProxyValue } from '@sourcegraph/comlink'
import { Subject } from 'rxjs'
import { TextDocument } from 'sourcegraph'
import { TextModel } from '../../client/services/modelService'
import { ExtDocument } from './textDocument'

/** @internal */
export interface ExtDocumentsAPI extends ProxyValue {
    $acceptDocumentData(models: readonly TextModel[]): void
}

/** @internal */
export class ExtDocuments implements ExtDocumentsAPI, ProxyValue {
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

    public $acceptDocumentData(models: readonly TextModel[]): void {
        for (const model of models) {
            const isNew = !this.documents.has(model.uri)
            const doc = new ExtDocument(model)
            this.documents.set(model.uri, doc)
            if (isNew) {
                console.log('ISNEW', doc)
                this.openedTextDocuments.next(doc)
                console.log('ISNEW done')
            }
        }
    }
}
