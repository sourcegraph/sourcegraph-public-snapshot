import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import * as clientType from '@sourcegraph/extension-api-types'
import { Subject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { URI } from '../types/uri'

/** @internal */
export interface ExtRootsAPI extends ProxyValue {
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

    public readonly changes = new Subject<void>()

    public $acceptRoots(roots: clientType.WorkspaceRoot[]): void {
        this.roots = Object.freeze(roots.map(plain => ({ ...plain, uri: new URI(plain.uri) })))
        this.changes.next()
    }
}
