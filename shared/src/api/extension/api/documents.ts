import { ProxyResult, ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { Subject } from 'rxjs'
import { TextDocument } from 'sourcegraph'
import { ClientDocumentsAPI } from '../../client/api/documents'
import { TextModel } from '../../client/services/modelService'
import { ExtDocument } from './textDocument'

/** @internal */
export interface ExtDocumentsAPI extends ProxyValue {
    $acceptDocumentData(models: readonly TextModel[]): void
}

/** @internal */
export class ExtDocuments implements ExtDocumentsAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private documents = new Map<string, ExtDocument>()

    constructor(private proxy: ProxyResult<ClientDocumentsAPI>, private sync: () => Promise<void>) {}

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
            const doc = this.addDocument(model)
            if (isNew) {
                this.openedTextDocuments.next(doc)
            }
        }
    }

    public async openTextDocument(uri: URL): Promise<ExtDocument> {
        const uriStr = uri.toString()
        const doc = this.documents.get(uriStr)
        if (doc) {
            return doc
        }
        const model = await this.proxy.$openTextDocument(uri.toString())
        return this.addDocument(model)
    }

    private addDocument(model: TextModel): ExtDocument {
        const doc = new ExtDocument(model)
        this.documents.set(model.uri, doc)
        return doc
    }
}
