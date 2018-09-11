import { Observable, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import { TextDocument } from 'sourcegraph'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtDocumentsAPI } from '../../extension/api/documents'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { TextDocumentItem } from '../types/textDocument'
import { SubscriptionMap } from './common'

/** @internal */
export class ClientDocuments {
    private subscriptions = new Subscription()
    private registrations = new SubscriptionMap()
    private proxy: ExtDocumentsAPI

    constructor(
        connection: Connection,
        environmentTextDocument: Observable<Pick<TextDocument, 'uri' | 'languageId'> | null>
    ) {
        this.proxy = createProxyAndHandleRequests('documents', connection, this)

        this.subscriptions.add(
            environmentTextDocument.pipe(filter((v): v is TextDocumentItem => v !== null)).subscribe(doc => {
                this.proxy.$acceptDocumentData(doc)
            })
        )

        this.subscriptions.add(this.registrations)
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
