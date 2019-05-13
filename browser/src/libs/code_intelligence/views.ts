import { from, merge, Observable } from 'rxjs'
import { concatAll, filter, mergeMap } from 'rxjs/operators'
import { isInstanceOf } from '../../../../shared/src/util/types'
import { MutationRecordLike } from '../../shared/util/dom'

/**
 * An addition or removal of a view observed by a {@link MutationObserver}.
 *
 * @template V The type of view, such as a code view.
 */
export interface ViewEvent {
    type: 'added' | 'removed'
    element: HTMLElement
}

/**
 * Finds and resolves elements matched by a MutationObserver to views.
 *
 * @template V The type of view, such as a code view.
 */
export type ViewResolver<V = {}> = (container: Element) => Iterable<V & Pick<ViewEvent, 'element'>>

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
    viewResolvers: ViewResolver<V & Pick<ViewEvent, 'element'>>[]
): (mutations: Observable<MutationRecordLike[]>) => Observable<V & ViewEvent> {
    const observeViewEvents = (
        nodeList: ArrayLike<Node> & Iterable<Node>,
        type: ViewEvent['type']
    ): Observable<V & ViewEvent> =>
        from(nodeList).pipe(
            filter(isInstanceOf(HTMLElement)),
            mergeMap(addedElement =>
                from(viewResolvers).pipe(
                    mergeMap(resolveViews => [...resolveViews(addedElement)].map(view => ({ ...view, type })))
                )
            )
        )
    return mutations =>
        mutations.pipe(
            concatAll(),
            mergeMap(mutation =>
                merge(
                    observeViewEvents(mutation.addedNodes, 'added'),
                    observeViewEvents(mutation.removedNodes, 'removed')
                )
            )
        )
}
