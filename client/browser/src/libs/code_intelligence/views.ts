import { from, merge, Observable } from 'rxjs'
import { concatAll, filter, map, mergeMap } from 'rxjs/operators'
import { isDefined, isInstanceOf } from '../../../../../shared/src/util/types'
import { MutationRecordLike, querySelectorAllOrSelf } from '../../shared/util/dom'

/**
 * Finds and resolves elements matched by a MutationObserver to views.
 *
 * @template V The type of view, such as a code view.
 */
export interface ViewResolver<V> {
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

interface AddedViewEvent {
    type: 'added'
}

interface RemovedView {
    type: 'removed'

    /** The HTML element that was removed. */
    element: HTMLElement
}

/**
 * An addition or removal of a view observed by a {@link MutationObserver}.
 *
 * @template V The type of view, such as a code view.
 */
export type ViewEvent<V> = (AddedViewEvent & V) | RemovedView

/**
 * Find all the views (e.g., code views) on a page using view resolvers (defined in
 * {@link CodeHost}).
 *
 * Emits every view that gets added or removed.
 *
 * At any given time, there can be any number of views on a page.
 *
 * @template V The type of view, such as a code view.
 */
export function trackViews<V>(
    viewResolvers: ViewResolver<V>[]
): (mutations: Observable<MutationRecordLike[]>) => Observable<ViewEvent<V>> {
    return mutations =>
        mutations.pipe(
            concatAll(),
            mergeMap(mutation =>
                merge(
                    // Find all new code views within the added nodes
                    // (MutationObservers don't emit all descendant nodes of an addded node recursively)
                    from(mutation.addedNodes).pipe(
                        filter(isInstanceOf(HTMLElement)),
                        mergeMap(addedElement =>
                            from(viewResolvers).pipe(
                                mergeMap(spec =>
                                    [...(querySelectorAllOrSelf(addedElement, spec.selector) as Iterable<HTMLElement>)]
                                        .map(element => {
                                            const view = spec.resolveView(element)
                                            return (
                                                view && {
                                                    ...view,
                                                    type: 'added' as const,
                                                }
                                            )
                                        })
                                        .filter(isDefined)
                                )
                            )
                        )
                    ),
                    // For removed nodes, find the removed elements, but don't resolve the kind (it's not relevant)
                    from(mutation.removedNodes).pipe(
                        filter(isInstanceOf(HTMLElement)),
                        mergeMap(removedElement =>
                            from(viewResolvers).pipe(
                                mergeMap(
                                    ({ selector }) =>
                                        querySelectorAllOrSelf(removedElement, selector) as Iterable<HTMLElement>
                                ),
                                map(element => ({
                                    element,
                                    type: 'removed' as const,
                                }))
                            )
                        )
                    )
                )
            )
        )
}
