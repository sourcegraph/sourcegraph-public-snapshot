import { ProxyResult } from '@sourcegraph/comlink'
import { Subscription } from 'rxjs'
import { ExtRootsAPI } from '../../extension/api/roots'
import { WorkspaceService } from '../services/workspaceService'

/** @internal */
export class ClientRoots {
    private subscriptions = new Subscription()

    constructor(proxy: ProxyResult<ExtRootsAPI>, workspaceService: WorkspaceService) {
        this.subscriptions.add(
            workspaceService.roots.subscribe(roots => {
                // tslint:disable-next-line: no-floating-promises
                proxy.$acceptRoots(roots || [])
            })
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
