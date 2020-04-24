import { Remote } from '@sourcegraph/comlink'
import { Subscription } from 'rxjs'
import { ExtRootsAPI } from '../../extension/api/roots'
import { WorkspaceService } from '../services/workspaceService'

/** @internal */
export class ClientRoots {
    private subscriptions = new Subscription()

    constructor(proxy: Remote<ExtRootsAPI>, workspaceService: WorkspaceService) {
        this.subscriptions.add(
            workspaceService.roots.subscribe(roots => {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                proxy.$acceptRoots(roots || [])
            })
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
