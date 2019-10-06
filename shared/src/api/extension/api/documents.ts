import { ProxyValue, proxyValueSymbol, ProxyResult } from '@sourcegraph/comlink'
import { Subject } from 'rxjs'
import { TextDocument } from 'sourcegraph'
import { TextModelUpdate, TextModel } from '../../client/services/modelService'
import { ExtDocument } from './textDocument'
import { ClientDocumentsAPI } from '../../client/api/documents'

/** @internal */
export interface ExtDocumentsAPI extends ProxyValue {
    $acceptDocumentData(modelUpdates: readonly TextModelUpdate[]): void
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

    public $acceptDocumentData(modelUpdates: readonly TextModelUpdate[]): void {
        for (const update of modelUpdates) {
            switch (update.type) {
                case 'added': {
                    const { uri, text, languageId }: TextModel = update
                    const doc = this.addDocument({ uri, text, languageId })
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
