import { BehaviorSubject, Observable, Unsubscribable } from 'rxjs'
import { map } from 'rxjs/operators'

/** A registry entry for a registered provider. */
export interface Entry<O, P> {
    registrationOptions: O
    provider: P
}

/** Base class for provider registries for features. */
export abstract class FeatureProviderRegistry<O, P> {
    protected entries = new BehaviorSubject<Entry<O, P>[]>([])

    public constructor(initialEntries?: Entry<O, P>[]) {
        if (initialEntries) {
            this.entries.next(initialEntries)
        }
    }

    public registerProvider(registrationOptions: O, provider: P): Unsubscribable {
        const entry: Entry<O, P> = { registrationOptions, provider }
        this.entries.next([...this.entries.value, entry])
        return {
            unsubscribe: () => {
                this.entries.next(this.entries.value.filter(e => e !== entry))
            },
        }
    }

    /** All providers, emitted whenever the set of registered providers changed. */
    public readonly providers: Observable<P[]> = this.entries.pipe(
        map(entries => entries.map(({ provider }) => provider))
    )
}
