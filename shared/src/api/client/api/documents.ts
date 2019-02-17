import { ProxyResult } from 'comlink'
import { Observable, Subscription } from 'rxjs'
import { ExtDocumentsAPI } from '../../extension/api/documents'
import { TextDocumentItem } from '../types/textDocument'

/** @internal */
export class ClientDocuments {
    private subscriptions = new Subscription()

    constructor(
        private proxy: ProxyResult<ExtDocumentsAPI>,
        modelTextDocuments: Observable<TextDocumentItem[] | null>
    ) {
        this.subscriptions.add(
            modelTextDocuments.subscribe(docs => {
                // tslint:disable-next-line: no-floating-promises
                this.proxy.$acceptDocumentData(docs || [])
            })
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
