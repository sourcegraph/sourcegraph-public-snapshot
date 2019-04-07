import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
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
    public readonly [proxyValueSymbol] = true

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
        // This 2nd sync is necessary after the monolithic model was split into ModelService and
        // EditorService, which means that changes to the model/editor state are not atomic. Without
        // the 2nd sync, the `document not found: ...` error above is thrown when the user navigates
        // between files (e.g., using go-to-definition) because the hover is requested on the
        // destination file before it has been added (because there are now more "steps" to adding
        // the file: first clearing the current model and editor and then adding the new model and
        // editor).
        //
        // TODO: add an atomic way to update the state to remove this hack
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
                this.openedTextDocuments.next(doc)
            }
        }
    }
}
