import { Location } from '@sourcegraph/extension-api-types'
import { from, Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { match, TextDocumentIdentifier } from '../types/textDocument'
import { CodeEditorWithPartialModel } from './editorService'
import { DocumentFeatureProviderRegistry } from './registry'
import { flattenAndCompact } from './util'

/**
 * Function signature for retrieving related locations given a location (e.g., definition, implementation, and type
 * definition).
 */
export type ProvideTextDocumentLocationSignature<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
> = (params: P) => Observable<L[] | null>

/**
 * Provides location results from matching registered providers for definition, implementation, and type definition
 * requests.
 */
export class TextDocumentLocationProviderRegistry<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
> extends DocumentFeatureProviderRegistry<ProvideTextDocumentLocationSignature<P, L>> {
    /**
     * Returns an observable that, initially and upon the registered provider set changing, emits an
     * observable of the combined location results from all providers.
     *
     * Using a higher-order observable here lets the caller display a loading indicator when some
     * providers have not yet completed (i.e., when the inner observable is not yet completed). The
     * outer observable never completes because providers may be registered and unregistered at any
     * time.
     */
    public getLocations(params: P): Observable<Observable<L[] | null>> {
        return getLocationsFromProviders(this.providersForDocument(params.textDocument), params)
    }

    /**
     * Reports whether there are any location providers registered for the active text document.
     * This can be used, for example, to selectively show a "Find references" button if there are
     * any reference providers registered.
     *
     * @param editors The code editors in {@link EditorService}.
     */
    public hasProvidersForActiveTextDocument(editors: readonly CodeEditorWithPartialModel[]): Observable<boolean> {
        return this.entries.pipe(
            map(entries => {
                const activeEditor = editors.find(({ isActive }) => isActive)
                if (!activeEditor) {
                    return false
                }
                return (
                    entries.filter(({ registrationOptions }) =>
                        match(registrationOptions.documentSelector, {
                            uri: activeEditor.resource,
                            languageId: activeEditor.model.languageId,
                        })
                    ).length > 0
                )
            })
        )
    }
}

/**
 * Registration options for a text document provider that has an ID (such as {@link sourcegraph.LocationProvider}).
 */
export interface TextDocumentProviderIDRegistrationOptions extends TextDocumentRegistrationOptions {
    /**
     * The identifier of the provider, used to distinguish it among other providers.
     *
     * This corresponds to, e.g., the `id` parameter in {@link sourcegraph.languages.registerLocationProvider}.
     */
    id: string
}

/**
 * The registry for text document location providers with a distinguishing ID (i.e., registered using
 * {@link TextDocumentProviderIDRegistrationOptions}).
 */
export class TextDocumentLocationProviderIDRegistry extends DocumentFeatureProviderRegistry<
    ProvideTextDocumentLocationSignature<TextDocumentPositionParams, Location>,
    TextDocumentProviderIDRegistrationOptions
> {
    /**
     * @param id The provider ID.
     * @returns an observable of the set of registered providers that apply to the document with the given ID.
     * (Usually there is at most 1 such provider.) The observable emits initially and whenever the set changes (due
     * to a provider being registered or unregistered).
     */
    public providersForDocumentWithID(
        id: string,
        document: TextDocumentIdentifier
    ): Observable<ProvideTextDocumentLocationSignature<TextDocumentPositionParams, Location>[]> {
        return this.providersForDocument(document, registrationOptions => registrationOptions.id === id)
    }

    /**
     * Gets locations from the provider with the given ID (i.e., the `id` parameter to
     * {@link sourcegraph.languageFeatures.registerLocationProvider}). Returns an observable that,
     * initially and upon the provider changing, emits an observable of the provider's location
     * results.
     *
     * Using a higher-order observable here lets the caller display a loading indicator when the
     * inner observable has not yet completed. The outer observable never completes because
     * providers may be registered and unregistered at any time.
     *
     * @param id The provider ID.
     */
    public getLocations(id: string, params: TextDocumentPositionParams): Observable<Observable<Location[] | null>> {
        return getLocationsFromProviders(this.providersForDocumentWithID(id, params.textDocument), params)
    }
}

/**
 * Returns the combined results of invoking multiple location providers.
 *
 * @internal Callers should instead use the the getLocations or similarly named methods on classes
 * defined in this module.
 */
export function getLocationsFromProviders<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
>(
    providers: Observable<ProvideTextDocumentLocationSignature<P, L>[]>,
    params: P,
    logErrors = true
): Observable<Observable<L[] | null>> {
    return providers.pipe(
        map(providers =>
            combineLatestOrDefault(
                providers.map(provider =>
                    from(provider(params)).pipe(
                        catchError(err => {
                            if (logErrors) {
                                console.error(err)
                            }
                            return [null]
                        })
                    )
                )
            ).pipe(map(locations => flattenAndCompact(locations)))
        )
    )
}
