import { RefObject, useEffect, useState } from 'react'

import { ObservableInput } from 'rxjs'

import { LazyQueryState, useLazyParallelRequest } from '../../../hooks/use-parallel-requests/use-parallel-request'

export interface UseInsightDataResult<T> {
    isVisible: boolean
    query: (request: () => ObservableInput<T>) => void
    state: LazyQueryState<T>
}

/**
 * This hook runs consumer request function in parallel with other requests that have
 * been made by useInsightData hook if and only if the element {@link reference} prop
 * is visible on the screen.
 *
 * @param request - consumer's request to run
 * @param reference - consumer's element that should be visible to run consumer's request
 */
export function useInsightData<D>(
    request: () => ObservableInput<D>,
    reference: RefObject<HTMLElement>
): UseInsightDataResult<D> {
    const { state, query } = useLazyParallelRequest<D>()

    const [isVisible, setVisibility] = useState<boolean>(false)
    const [hasIntersected, setHasIntersected] = useState<boolean>(false)

    useEffect(() => {
        if (hasIntersected) {
            // eslint-disable-next-line @typescript-eslint/unbound-method
            const { unsubscribe } = query(request)

            return unsubscribe
        }

        return
    }, [hasIntersected, query, request])

    useEffect(() => {
        const element = reference.current

        // Do not observe insights visibility for non GQL based APIs.
        // Only GQL API supports partial insights fetching based on
        // insights visibility.
        if (!element) {
            return
        }

        function handleIntersection(entries: IntersectionObserverEntry[]): void {
            const [entry] = entries

            setVisibility(entry.isIntersecting)

            if (entry.isIntersecting) {
                setHasIntersected(true)
            }
        }

        const observer = new IntersectionObserver(handleIntersection)

        observer.observe(element)

        return () => observer.unobserve(element)
    }, [reference])

    return { state, isVisible, query }
}
