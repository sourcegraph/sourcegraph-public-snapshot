import { BehaviorSubject, Observable, TeardownLogic } from 'rxjs'
import { first, map } from 'rxjs/operators'
import { TextDocumentRegistrationOptions } from '../../protocol'

interface Entry<O extends TextDocumentRegistrationOptions, P> {
    registrationOptions: O
    provider: P
}

/** Base class for provider registries for text document features. */
export abstract class TextDocumentFeatureProviderRegistry<O extends TextDocumentRegistrationOptions, P> {
    private entries = new BehaviorSubject<Entry<O, P>[]>([])

    public registerProvider(registrationOptions: O, provider: P): TeardownLogic {
        const entry: Entry<O, P> = { registrationOptions, provider }
        this.entries.next([...this.entries.value, entry])
        return () => {
            this.entries.next(this.entries.value.filter(e => e !== entry))
        }
    }

    /** All providers, emitted whenever the set of registered providers changed. */
    public readonly providers: Observable<P[]> = this.entries.pipe(
        map(providers => providers.map(({ provider }) => provider))
    )

    /**
     * The current set of providers. Used by callers that do not need to react to providers being registered or
     * unregistered.
     */
    public readonly providersSnapshot = this.providers.pipe(first())
}

/** An empty provider registry, mainly useful in tests and example code. */
export class NoopProviderRegistry extends TextDocumentFeatureProviderRegistry<any, any> {}
