import { Observable, Subscription } from 'rxjs'
import { TextDocument } from 'sourcegraph'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtDocumentsAPI } from '../../extension/api/documents'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { SubscriptionMap } from './common'

/** @internal */
export class ClientDocuments {
    private subscriptions = new Subscription()
    private registrations = new SubscriptionMap()
    private proxy: ExtDocumentsAPI

    constructor(
        connection: Connection,
        environmentTextDocuments: Observable<Pick<TextDocument, 'uri' | 'languageId' | 'text'>[] | null>
    ) {
        this.proxy = createProxyAndHandleRequests('documents', connection, this)

        this.subscriptions.add(
            environmentTextDocuments.subscribe(docs => {
                this.proxy.$acceptDocumentData(docs || [])
            })
        )

        this.subscriptions.add(this.registrations)
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
