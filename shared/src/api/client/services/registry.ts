import { BehaviorSubject, Observable, Unsubscribable } from 'rxjs'
import { map } from 'rxjs/operators'
import { getModeFromPath } from '../../../languages'
import { TextDocumentRegistrationOptions } from '../../protocol'
import { match, TextDocumentIdentifier } from '../types/textDocument'

/** A registry entry for a registered provider. */
export interface Entry<O, P> {
    registrationOptions: O
    provider: P
}

/** Base class for provider registries for features. */
export abstract class FeatureProviderRegistry<O, P> {
    protected entries = new BehaviorSubject<Entry<O, P>[]>([])

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

/**
 * A registry for providers that provide features within a document. Use this class instead of
 * {@link FeatureProviderRegistry} when all calls to the provider are scoped to a document.
 *
 * For example, hovers are scoped to a document (i.e., the document URI is one of the required arguments passed to
 * the hover provider), so this class is used for the hover provider registry.
 */
export abstract class DocumentFeatureProviderRegistry<P> extends FeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    P
> {
    /**
     * Returns an observable of the set if registered providers that apply to the document. The observable emits
     * initially and whenever the set changes (due to a provider being registered or unregistered).
     */
    public providersForDocument(document: TextDocumentIdentifier): Observable<P[]> {
        return this.entries.pipe(
            map(entries =>
                entries
                    .filter(({ registrationOptions }) =>
                        match(registrationOptions.documentSelector, {
                            uri: document.uri,
                            languageId: getModeFromPath(document.uri),
                        })
                    )
                    .map(({ provider }) => provider)
            )
        )
    }
}
