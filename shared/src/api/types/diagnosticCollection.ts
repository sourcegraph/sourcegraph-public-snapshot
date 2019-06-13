import { Diagnostic } from '@sourcegraph/extension-api-types'
import { Subject, Subscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

/**
 * The severity of a diagnostic.
 *
 * This is needed because if sourcegraph.DiagnosticSeverity enum values are referenced, the
 * `sourcegraph` module import at the top of the file is emitted in the generated code.
 */
export const DiagnosticSeverity: typeof sourcegraph.DiagnosticSeverity = {
    Error: 0,
    Warning: 1,
    Information: 2,
    Hint: 3,
}

/**
 * A collection of diagnostics.
 */
export class DiagnosticCollection<D extends Diagnostic> {
    /** Map of resource URI to the resource's diagnostics. */
    private data = new Map<string, D[]>()

    private _changes = new Subject<URL[]>()

    constructor(public readonly name: string) {}

    public set(uri: URL | string, diagnostics: D[] | undefined, merge?: boolean, emitChanged?: boolean): void
    public set(entries: [URL | string, D[] | undefined][]): void
    public set(
        arg1: URL | string | [URL | string, D[] | undefined][],
        arg2?: D[] | undefined,
        merge = false,
        emitChanged = true
    ): void {
        if (Array.isArray(arg1)) {
            const beforeUris = Array.from(this.data.keys())
            this.clear(false)
            for (const [uri, diagnostics] of arg1) {
                this.set(uri.toString(), diagnostics, true, false)
            }
            this.changed([...beforeUris, ...arg1.map(([uri]) => uri.toString())])
        } else {
            const key = arg1.toString()
            if (arg2) {
                this.data.set(key, merge ? [...(this.data.get(key) || []), ...arg2] : arg2)
            } else {
                this.data.delete(key)
            }
            if (emitChanged) {
                this.changed(arg1)
            }
        }
    }

    public delete(uri: URL | string): void {
        this.data.delete(uri.toString())
        this.changed(uri)
    }

    public clear(emitChanged = true): void {
        const uris = [...this.data.keys()]
        this.data.clear()
        if (emitChanged) {
            this.changed(uris)
        }
    }

    private changed(uris: URL | string | (URL | string)[]): void {
        if (Array.isArray(uris) && uris.length === 0) {
            return
        }
        this._changes.next((Array.isArray(uris) ? uris : [uris]).map(u => (typeof u === 'string' ? new URL(u) : u)))
    }

    public readonly changes: Subscribable<URL[]> = this._changes

    public *entries(): IterableIterator<[URL, D[]]> {
        for (const [uri, diagnostics] of this.data) {
            yield [new URL(uri), diagnostics]
        }
    }

    public get(uri: URL | string): readonly D[] | undefined {
        return this.data.get(uri.toString())
    }

    public has(uri: URL | string): boolean {
        return this.data.has(uri.toString())
    }

    public unsubscribe(): void {
        this.clear()
        this._changes.unsubscribe()
    }
}

/**
 * A diagnostic collection that can only be read from, not written to. Other callers may write to
 * this diagnostic collection.
 */
export interface ReadonlyDiagnosticCollection
    extends Pick<DiagnosticCollection<Diagnostic>, 'changes' | 'entries' | 'get' | 'has'> {}
