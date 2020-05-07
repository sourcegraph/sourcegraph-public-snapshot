import { Observable, BehaviorSubject, of, combineLatest, concat, Subscription } from 'rxjs'
import { View as ExtensionView, DirectoryViewContext } from 'sourcegraph'
import { switchMap, map, distinctUntilChanged, startWith, delay, catchError } from 'rxjs/operators'
import { Evaluated, Contributions, ContributableViewContainer } from '../../protocol'
import { isEqual } from 'lodash'
import { asError, ErrorLike } from '../../../util/errors'
import { finallyReleaseProxy } from '../api/common'
import { DeepReplace, isNot, isExactly } from '../../../util/types'

/**
 * A view is a page or partial page.
 */
export interface View extends ExtensionView {}

/**
 * A map from type of container names to the internal type of the context parameter provided by the container.
 */
export interface ViewContexts {
    [ContributableViewContainer.Panel]: never
    [ContributableViewContainer.GlobalPage]: Record<string, string>
    [ContributableViewContainer.Directory]: DeepReplace<DirectoryViewContext, URL, string>
}

export type ViewProviderFunction<C> = (context: C) => Observable<View | null>

/**
 * The view service manages views, which are pages or partial pages of content.
 */
export interface ViewService {
    /**
     * Register a view provider. Throws if the `id` is already registered.
     */
    register<W extends ContributableViewContainer>(
        id: string,
        where: W,
        provideView: ViewProviderFunction<ViewContexts[W]>
    ): Subscription

    /**
     * Get all providers for the given container.
     *
     * @todo return a Map by ID and make this the primary API
     */
    getWhere<W extends ContributableViewContainer>(where: W): Observable<ViewProviderFunction<ViewContexts[W]>[]>

    /**
     * Get a view's content. The returned observable emits whenever the content changes. If there is
     * no view with the given {@link id}, it emits `null`.
     */
    get<W extends ContributableViewContainer>(id: string, context: ViewContexts[W]): Observable<View | null>
}

/**
 * Creates a new {@link ViewService}.
 */
export const createViewService = (): ViewService => {
    interface Provider<W extends ContributableViewContainer> {
        where: W
        provideView: ViewProviderFunction<ViewContexts[W]>
    }
    const providers = new BehaviorSubject<Map<string, Provider<ContributableViewContainer>>>(new Map())

    return {
        register: <W extends ContributableViewContainer>(
            id: string,
            where: W,
            provideView: ViewProviderFunction<ViewContexts[W]>
        ) => {
            if (providers.value.has(id)) {
                throw new Error(`view already exists with ID ${id}`)
            }
            const provider = { where, provideView }
            providers.value.set(id, provider as any) // TODO: find a type-safe way
            providers.next(providers.value)
            return new Subscription(() => {
                const p = providers.value.get(id)
                if (p?.provideView === provideView) {
                    // Check equality to ensure we only unsubscribe the exact same provider we
                    // registered, not some other provider that was registered later with the same
                    // ID.
                    providers.value.delete(id)
                    providers.next(providers.value)
                }
            })
        },
        getWhere: where =>
            providers.pipe(
                map(providers => [...providers.values()].filter(e => e.where === where).map(e => e.provideView)),
                distinctUntilChanged((a, b) => isEqual(a, b))
            ),
        get: (id, context) =>
            providers.pipe(
                map(providers => providers.get(id)?.provideView),
                distinctUntilChanged(),
                switchMap(provider => (provider ? provider(context).pipe(finallyReleaseProxy()) : of(null)))
            ),
    }
}

/**
 * Returns an observable of the view and its associated contents, `undefined` if loading, and `null`
 * if not found. The view must be both registered as a contribution and at runtime with the
 * {@link ViewService}.
 */
export const getView = <W extends ContributableViewContainer>(
    viewID: string,
    viewContainer: W,
    params: ViewContexts[W],
    contributions: Observable<Pick<Evaluated<Contributions>, 'views'>>,
    viewService: Pick<ViewService, 'get'>
): Observable<View | undefined | null> =>
    combineLatest([
        contributions
            .pipe(
                map(contributions =>
                    contributions.views?.some(({ id, where }) => id === viewID && where === viewContainer)
                )
            )
            .pipe(distinctUntilChanged()),
        viewService.get<W>(viewID, params).pipe(distinctUntilChanged()),

        // Wait for extensions to load for up to 5 seconds (grace period) before showing
        // "not found", to avoid showing an error for a brief period during initial
        // load.
        of(false).pipe(delay(5000), startWith(true)),
    ]).pipe(
        map(([isContributed, view, isInitialLoadGracePeriod]) =>
            isContributed && view ? view : isInitialLoadGracePeriod ? undefined : null
        ),
        distinctUntilChanged()
    )

export const getViewsForContainer = <W extends ContributableViewContainer>(
    where: W,
    params: ViewContexts[W],
    viewService: Pick<ViewService, 'getWhere'>
): Observable<(View | undefined | ErrorLike)[]> =>
    viewService.getWhere(where).pipe(
        switchMap(providers =>
            combineLatest([
                of(null), // don't block forever if no providers
                ...providers.map(provider =>
                    concat(
                        [undefined], // don't block other providers on first emission
                        provider(params).pipe(
                            catchError((err): [ErrorLike] => {
                                console.error('View provider errored:', err)
                                return [asError(err)]
                            })
                        )
                    )
                ),
            ])
        ),
        map(views => views.filter(isNot(isExactly(null))))
    )
