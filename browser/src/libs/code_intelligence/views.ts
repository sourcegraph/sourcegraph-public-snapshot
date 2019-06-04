import { animationFrameScheduler, from, Observable, Subscription } from 'rxjs'
import { concatAll, filter, mergeMap, observeOn, tap } from 'rxjs/operators'
import { isDefined, isInstanceOf } from '../../../../shared/src/util/types'
import { MutationRecordLike, querySelectorAllOrSelf } from '../../shared/util/dom'

interface View {
    element: HTMLElement
}

type ViewWithSubscriptions<V extends View> = V & {
    /**
     * Maintains subscriptions to resources that should be freed when the view is removed.
     */
    subscriptions: Subscription
}

/**
 * Finds and resolves elements matched by a MutationObserver to views.
 *
 * @template V The type of view, such as a code view.
 */
export interface ViewResolver<V extends View> {
    /**
     * The element selector (used with {@link Window#querySelectorAll}) that matches candidate
     * elements to be passed to {@link ViewResolver#resolveView}.
     */
    selector: string

    /**
     * Resolve an element matched by {@link ViewResolver#selector} to a view, or `null` if it's not
     * a a valid view upon further examination.
     */
    resolveView: (element: HTMLElement) => V | null
}

/**
 * Find all the views (e.g., code views) on a page using view resolvers (defined in
 * {@link CodeHost}).
 *
 * Emits every view that gets added as a {@link ViewWithSubscriptions},
 * and frees a view's resources when it gets removed from the page.
 *
 * At any given time, there can be any number of views on a page.
 *
 * @template V The type of view, such as a code view.
 */
export function trackViews<V extends View>(
    viewResolvers: ViewResolver<V>[]
): (mutations: Observable<MutationRecordLike[]>) => Observable<ViewWithSubscriptions<V>> {
    const viewStates = new Map<HTMLElement, ViewWithSubscriptions<V>>()
    return mutations =>
        mutations.pipe(
            observeOn(animationFrameScheduler),
            concatAll(),
            // Inspect removed nodes for known views
            tap(({ removedNodes }) => {
                for (const node of removedNodes) {
                    if (!(node instanceof HTMLElement)) {
                        continue
                    }
                    const view = viewStates.get(node)
                    if (view) {
                        view.subscriptions.unsubscribe()
                        viewStates.delete(node)
                        continue
                    }
                    for (const viewElement of viewStates.keys()) {
                        if (node.contains(viewElement)) {
                            viewStates.get(viewElement)!.subscriptions.unsubscribe()
                            viewStates.delete(viewElement)
                        }
                    }
                }
            }),
            mergeMap(mutation =>
                // Find all new code views within the added nodes
                // (MutationObservers don't emit all descendant nodes of an addded node recursively)
                from(mutation.addedNodes).pipe(
                    filter(isInstanceOf(HTMLElement)),
                    mergeMap(addedElement =>
                        from(viewResolvers).pipe(
                            mergeMap(({ selector, resolveView }) =>
                                [...querySelectorAllOrSelf<HTMLElement>(addedElement, selector)].map(
                                    (element): ViewWithSubscriptions<V> | null => {
                                        const view = resolveView(element)
                                        return (
                                            view && {
                                                ...view,
                                                subscriptions: new Subscription(),
                                            }
                                        )
                                    }
                                )
                            ),
                            filter(
                                (view): view is ViewWithSubscriptions<V> =>
                                    isDefined(view) && !viewStates.has(view.element)
                            ),
                            tap(view => {
                                viewStates.set(view.element, view)
                            })
                        )
                    )
                )
            )
        )
}
