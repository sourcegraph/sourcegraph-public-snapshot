import { ProxyResult, ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { Range } from '@sourcegraph/extension-api-classes'
import { from, Unsubscribable } from 'rxjs'
import { map } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ClientDiagnosticsAPI, DiagnosticData } from '../../client/api/diagnostics'
import { DiagnosticCollection } from '../../types/diagnosticCollection'
import { fromDiagnostic } from './types'

/** @internal */
export interface ExtDiagnosticsAPI extends ProxyValue {
    // TODO!(sqs): inefficient
    $acceptDiagnosticData(updates: DiagnosticData): void
}

class DiagnosticCollectionWithUnsubscribeCallback extends DiagnosticCollection<sourcegraph.Diagnostic> {
    public onUnsubscribe?: () => void

    public unsubscribe(): void {
        if (this.onUnsubscribe) {
            this.onUnsubscribe()
        }
        super.unsubscribe()
    }
}

/** @internal */
// TODO!(sqs): this is weird because it stores duplicates of the diagnostics data on the ext host,
// one for the version of the data received roundtrip from the client and one the original
// DiagnosticCollection owned by an extension. See skipped test `$acceptDiagnosticData`.
export class ExtDiagnostics
    implements
        ExtDiagnosticsAPI,
        Pick<typeof sourcegraph.languages, 'diagnosticsChanges' | 'getDiagnostics' | 'createDiagnosticCollection'>,
        Unsubscribable {
    public readonly [proxyValueSymbol] = true

    /** All diagnostics data, from the client. */
    private data = new DiagnosticCollection<sourcegraph.Diagnostic>('')

    /** All diagnostic collections created on the extension host. */
    private collections: sourcegraph.DiagnosticCollection[] = []

    constructor(private proxy: ProxyResult<ClientDiagnosticsAPI>) {}

    public readonly diagnosticsChanges = from(this.data.changes).pipe(map(uris => ({ uris })))

    public $acceptDiagnosticData(data: DiagnosticData): void {
        this.data.set(
            data.map(([uri, diagnostics]) => [uri, diagnostics.map(d => ({ ...d, range: Range.fromPlain(d.range) }))])
        )
    }

    public getDiagnostics(resource: URL): sourcegraph.Diagnostic[]
    public getDiagnostics(): [URL, sourcegraph.Diagnostic[]][]
    public getDiagnostics(resource?: URL): sourcegraph.Diagnostic[] | [URL, sourcegraph.Diagnostic[]][] {
        if (resource) {
            const diagnostics: sourcegraph.Diagnostic[] = []
            for (const c of this.collections) {
                diagnostics.push(...(c.get(resource) || []))
            }
            return diagnostics
        }

        const merged = new Map<URL, sourcegraph.Diagnostic[]>()
        for (const c of this.collections) {
            for (const [uri, diagnostics] of c.entries()) {
                merged.set(uri, [...(merged.get(uri) || []), ...diagnostics])
            }
        }
        return [...merged.entries()]
    }

    public createDiagnosticCollection(name: string): sourcegraph.DiagnosticCollection {
        const c = new DiagnosticCollectionWithUnsubscribeCallback(name)
        this.collections.push(c)

        // Send the new data (from all collections) to the client when there is a change to any
        // collection.
        const subscription = c.changes.subscribe(() =>
            this.proxy.$acceptDiagnosticData(
                this.getDiagnostics().map(([uri, diagnostics]) => [uri.toString(), diagnostics.map(fromDiagnostic)])
            )
        )

        // Remove from ExtDiagnostics#collections array when the DiagnosticCollection is
        // unsubscribed.
        c.onUnsubscribe = () => {
            subscription.unsubscribe()
            const i = this.collections.indexOf(c)
            if (i !== -1) {
                this.collections.splice(i, 1)
            }
        }

        return c
    }

    public unsubscribe(): void {
        for (const c of this.collections) {
            c.unsubscribe()
        }
        this.data.unsubscribe()
    }
}
