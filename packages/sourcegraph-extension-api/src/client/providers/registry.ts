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
        map(providers => providers.map(({ provider }) => provider))
    )

    /**
     * The current set of providers. Used by callers that do not need to react to providers being
     * registered or unregistered.
     *
     * NOTE: You should usually use the providers property on this class, not providersSnapshot,
     * even when you think you don't need live-updating results. Providers are registered
     * asynchronously after the client connects (or reconnects) to the server. So, the providers
     * list might be empty at the instant you need the results (because the client was just
     * instantiated and is waiting on a network roundtrip before it registers providers, or because
     * there was a temporary network error and the client is reestablishing the connection). By
     * using the providers property, the consumer will get the results it (probably) expects when
     * the client connects and registers the providers.
     */
    public get providersSnapshot(): P[] {
        return this.entries.value.map(({ provider }) => provider)
    }
}
