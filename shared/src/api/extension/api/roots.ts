import * as clientType from '@sourcegraph/extension-api-types'
import { ProxyValue, proxyValueSymbol } from 'comlink'
import { Observable, Subject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { URI } from '../types/uri'

/** @internal */
export interface ExtRootsAPI {
    $acceptRoots(roots: clientType.WorkspaceRoot[]): void
}

/** @internal */
export class ExtRoots implements ExtRootsAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private roots: ReadonlyArray<sourcegraph.WorkspaceRoot> = []

    /**
     * Returns all workspace roots.
     *
     * @internal
     */
    public getAll(): ReadonlyArray<sourcegraph.WorkspaceRoot> {
        return this.roots
    }

    private changes = new Subject<void>()
    public readonly onDidChange: Observable<void> = this.changes

    public $acceptRoots(roots: clientType.WorkspaceRoot[]): void {
        this.roots = Object.freeze(
            roots.map(plain => ({ ...plain, uri: new URI(plain.uri) } as sourcegraph.WorkspaceRoot))
        )
        this.changes.next()
    }
}
