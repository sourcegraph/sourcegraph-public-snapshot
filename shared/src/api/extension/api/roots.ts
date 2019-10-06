import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import * as clientType from '@sourcegraph/extension-api-types'
import { Subject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

/** @internal */
export interface ExtRootsAPI extends ProxyValue {
    $acceptRoots(roots: readonly clientType.WorkspaceRoot[]): void
}

/** @internal */
export class ExtRoots implements ExtRootsAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private roots: readonly sourcegraph.WorkspaceRoot[] = []

    /**
     * Returns all workspace roots.
     *
     * @internal
     */
    public getAll(): readonly sourcegraph.WorkspaceRoot[] {
        return this.roots
    }

    public readonly changes = new Subject<void>()

    public $acceptRoots(roots: clientType.WorkspaceRoot[]): void {
        this.roots = Object.freeze(
            roots.map(plain => ({
                ...plain,
                uri: new URL(plain.uri),
                baseUri: plain.baseUri ? new URL(plain.baseUri) : undefined,
            }))
        )
        this.changes.next()
    }
}
