import { asyncScheduler, defer, from, Subscription, OperatorFunction, Subscribable, NextObserver, Subject } from 'rxjs'
import { concatAll, filter, mergeMap, observeOn, tap, mapTo, first } from 'rxjs/operators'
import { isDefined, isInstanceOf } from '../../../../shared/src/util/types'
import { MutationRecordLike, querySelectorAllOrSelf } from '../../shared/util/dom'

interface View {
    element: HTMLElement
}

export type ViewWithSubscriptions<V extends View> = V & {
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
 * Represents the state of a view tracked by {@link trackViews}.
 */
interface ViewState<V extends View> {
    /**
     * The view as emitted to consumers.
     */
    view: ViewWithSubscriptions<V>
    /**
     * Used to signal the view is close enough to the viewport to be emitted.
     */
    visibilitySpy: Subscribable<void> & NextObserver<void>
}

/**
 * Find all the views (e.g., code views) on a page using view resolvers (defined in
 * {@link CodeHost}).
 *
 * Emits every view that gets added and that is contained (even partially) in the viewport,
 * or within 4000px of the viewport's top/bottom edges.
 *
 * Views are emitted as {@link ViewWithSubscriptions}.
 * When a view gets removed from the page, its resources are freed.
 *
 * At any given time, there can be any number of views on a page.
 *
 * @template V The type of view, such as a code view.
 */
export function trackViews<V extends View>(
    viewResolvers: ViewResolver<V>[]
): OperatorFunction<MutationRecordLike[], ViewWithSubscriptions<V>> {
    return mutations =>
        defer(() => {
            const viewStates = new Map<HTMLElement, ViewState<V>>()

            /**
             * Observes code views, and emits once on {@link ViewState#visibilitySpy}
             * when a code view intersects with the specified bounding rectangle.
             */
            const intersectionObserver = new IntersectionObserver(
                (entries, observer) => {
                    for (const entry of entries) {
                        const viewState = viewStates.get(entry.target as HTMLElement)
                        if (!viewState || !entry.isIntersecting) {
                            continue
                        }
                        viewState.visibilitySpy.next()
                        observer.unobserve(entry.target)
                    }
                },
                { rootMargin: '4000px 0px' }
            )
            return mutations.pipe(
                observeOn(asyncScheduler),
                concatAll(),
                // Inspect removed nodes for known views
                tap(({ removedNodes }) => {
                    for (const node of removedNodes) {
                        if (!(node instanceof HTMLElement)) {
                            continue
                        }
                        const viewState = viewStates.get(node)
                        if (viewState) {
                            viewState.view.subscriptions.unsubscribe()
                            viewStates.delete(node)
                            continue
                        }
                        for (const [viewElement, { view }] of viewStates.entries()) {
                            if (node.contains(viewElement)) {
                                view.subscriptions.unsubscribe()
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
                                filter(isDefined),
                                filter(view => !viewStates.has(view.element)),
                                mergeMap(view => {
                                    const visibilitySpy = new Subject<void>()
                                    viewStates.set(view.element, {
                                        view,
                                        visibilitySpy,
                                    })
                                    // Observe changes to the visibility of the view's element,
                                    // and emit the view when the element is in or close enough to the viewport.
                                    intersectionObserver.observe(view.element)
                                    view.subscriptions.add(() => {
                                        intersectionObserver.unobserve(view.element)
                                    })
                                    return visibilitySpy.pipe(
                                        first(),
                                        mapTo(view)
                                    )
                                })
                            )
                        )
                    )
                )
            )
        })
}
