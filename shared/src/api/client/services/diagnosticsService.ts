import * as sourcegraph from 'sourcegraph'
import { Subscribable, BehaviorSubject } from 'rxjs'
import { map, distinctUntilChanged } from 'rxjs/operators'

/**
 * The diagnostics service publishes diagnostics about resources.
 */
export interface DiagnosticsService {
    /**
     * An observable that emits all diagnostics (from all collections) when any diagnostics change.
     */
    readonly all: Subscribable<[URL, sourcegraph.Diagnostic[]][]>

    /**
     * Observe a diagnostic collection's diagnostics.
     *
     * @param name The name of the diagnostic collection to observe.
     */
    observe(name: string): Subscribable<[URL, sourcegraph.Diagnostic[]][]>

    /**
     * Creates, updates, or deletes a diagnostic collection.
     *
     * @param name The name of the diagnostic collection to set.
     */
    set(name: string, data: [URL, sourcegraph.Diagnostic[]][] | null): void
}

/**
 * Creates a {@link DiagnosticsService} instance.
 */
export function createDiagnosticsService(): DiagnosticsService {
    const collections = new BehaviorSubject<{ [name: string]: [URL, sourcegraph.Diagnostic[]][] }>({})
    return {
        all: collections.pipe(
            map(collections => {
                const byUrl = new Map<string, sourcegraph.Diagnostic[]>()
                for (const [, diagnosticCollection] of Object.entries(collections)) {
                    for (const [uri, diagnostics] of diagnosticCollection) {
                        const key = uri.toString()
                        const existing = byUrl.get(key) || []
                        byUrl.set(key, [...existing, ...diagnostics])
                    }
                }
                return Array.from(byUrl.entries()).map<[URL, sourcegraph.Diagnostic[]]>(([uriStr, diagnostics]) => [
                    new URL(uriStr),
                    diagnostics,
                ])
            })
        ),
        observe: name =>
            collections.pipe(
                map(collections => collections[name] || []),
                distinctUntilChanged()
            ),
        set: (name, data) => {
            if (data) {
                collections.next({ ...collections.value, [name]: data })
            } else {
                const newCollections = { ...collections.value }
                delete newCollections[name]
                collections.next(newCollections)
            }
        },
    }
}
