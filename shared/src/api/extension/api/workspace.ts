import { ProxyMarked, proxyMarker } from 'comlink'
import * as clientType from '@sourcegraph/extension-api-types'
import { Subject, BehaviorSubject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

/** @internal */
export interface ExtWorkspaceAPI extends ProxyMarked {
    $acceptRoots(roots: readonly clientType.WorkspaceRoot[]): void
    $acceptVersionContext(versionContext: string | undefined): void
}

/** @internal */
export class ExtWorkspace implements ExtWorkspaceAPI, ProxyMarked {
    public readonly [proxyMarker] = true

    private roots: readonly sourcegraph.WorkspaceRoot[] = []

    /**
     * Returns all workspace roots.
     *
     * @internal
     */
    public getAllRoots(): readonly sourcegraph.WorkspaceRoot[] {
        return this.roots
    }

    public readonly rootChanges = new Subject<void>()
    public readonly versionContextChanges = new BehaviorSubject<string | undefined>(undefined)

    public $acceptRoots(roots: clientType.WorkspaceRoot[]): void {
        this.roots = Object.freeze(roots.map(plain => ({ ...plain, uri: new URL(plain.uri) })))
        this.rootChanges.next()
    }

    public $acceptVersionContext(versionContext: string | undefined): void {
        this.versionContextChanges.next(versionContext)
    }
}
