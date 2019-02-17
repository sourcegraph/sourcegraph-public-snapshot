import { WorkspaceRoot } from '@sourcegraph/extension-api-types'
import { ProxyResult } from 'comlink'
import { Observable, Subscription } from 'rxjs'
import { ExtRootsAPI } from '../../extension/api/roots'

/** @internal */
export class ClientRoots {
    private subscriptions = new Subscription()

    constructor(proxy: ProxyResult<ExtRootsAPI>, modelRoots: Observable<WorkspaceRoot[] | null>) {
        this.subscriptions.add(
            modelRoots.subscribe(roots => {
                // tslint:disable-next-line: no-floating-promises
                proxy.$acceptRoots(roots || [])
            })
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
