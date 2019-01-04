import { Observable, Subscription } from 'rxjs'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtDocumentsAPI } from '../../extension/api/documents'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { ViewComponentData } from '../model'
import { SubscriptionMap } from './common'

/** @internal */
export class ClientDocuments {
    private subscriptions = new Subscription()
    private registrations = new SubscriptionMap()
    private proxy: ExtDocumentsAPI

    constructor(connection: Connection, modelViewComponents: Observable<ViewComponentData[] | null>) {
        this.proxy = createProxyAndHandleRequests('documents', connection, this)

        this.subscriptions.add(
            modelViewComponents.subscribe(editors => {
                this.proxy.$acceptEditorData(editors || [])
            })
        )

        this.subscriptions.add(this.registrations)
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
