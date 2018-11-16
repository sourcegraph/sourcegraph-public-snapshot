import { Observable, Subscription } from 'rxjs'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtRootsAPI } from '../../extension/api/roots'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { WorkspaceRoot } from '../../protocol/plainTypes'

/** @internal */
export class ClientRoots {
    private subscriptions = new Subscription()
    private proxy: ExtRootsAPI

    constructor(connection: Connection, environmentRoots: Observable<WorkspaceRoot[] | null>) {
        this.proxy = createProxyAndHandleRequests('roots', connection, this)

        this.subscriptions.add(
            environmentRoots.subscribe(roots => {
                this.proxy.$acceptRoots(roots || [])
            })
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
