import { BehaviorSubject, Observable, Subject } from 'rxjs'
import { TextDocument } from 'sourcegraph'
import { ViewComponentData } from '../../client/model'
import { TextDocumentItem } from '../../client/types/textDocument'

/** @internal */
export interface ExtDocumentsAPI {
    $acceptEditorData(editors: ViewComponentData[]): void
}

/** @internal */
export class ExtDocuments implements ExtDocumentsAPI {
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
    public readonly onDidOpenTextDocument: Observable<TextDocument> = Observable.create(
        (observer: (doc: TextDocument) => void) => {
            console.warn(
                'workspace.onDidOpenTextDocument is deprecated and will soon be removed. Use workspace.activeTextDocument instead'
            )
            return this.textDocumentAdds.subscribe(observer)
        }
    )
    public readonly activeTextDocument = new BehaviorSubject<TextDocument | null>(null)

    public $acceptEditorData(editors: ViewComponentData[] | null): void {
        if (!editors || !editors.length) {
            this.activeTextDocument.next(null)
            return
        }
        for (const { item, isActive } of editors) {
            const isNew = !this.documents.has(item.uri)
            this.documents.set(item.uri, item)
            if (isNew) {
                this.textDocumentAdds.next(item)
            }
            if (isActive) {
                this.activeTextDocument.next(item)
            }
        }
    }
}
