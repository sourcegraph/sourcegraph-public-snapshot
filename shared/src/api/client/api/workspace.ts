import { Remote } from 'comlink'
import { Subscription } from 'rxjs'
import { ExtWorkspaceAPI } from '../../extension/api/workspace'
import { WorkspaceService } from '../services/workspaceService'

/** @internal */
export class ClientWorkspace {
    private subscriptions = new Subscription()

    constructor(proxy: Remote<ExtWorkspaceAPI>, workspaceService: WorkspaceService) {
        this.subscriptions.add(
            workspaceService.roots.subscribe(roots => {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                proxy.$acceptRoots(roots || [])
            })
        )

        this.subscriptions.add(
            workspaceService.versionContext.subscribe(versionContext => {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                proxy.$acceptVersionContext(versionContext)
            })
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
