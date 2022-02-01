import { useContext, useEffect, useState } from 'react'
import { ObservableInput } from 'rxjs'

import { ErrorLike } from '@sourcegraph/common'

import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from '../../../core/backend/gql-api/code-insights-gql-backend'
import { useLazyParallelRequest } from '../../../hooks/use-parallel-requests/use-parallel-request'

export interface UseInsightDataResult<T> {
    data: T | undefined
    error: ErrorLike | undefined
    loading: boolean
    isVisible: boolean
    query: (request: () => ObservableInput<T>) => void
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
    reference: React.RefObject<HTMLElement>
): UseInsightDataResult<D> {
    const api = useContext(CodeInsightsBackendContext)
    const isGqlAPI = api instanceof CodeInsightsGqlBackend

    const { data, loading, error, query } = useLazyParallelRequest<D>()

    // All non GQL API implementations do not support partial loading,
    // allowing insights fetching for these API whether insights are
    // in a viewport or not.
    const [isVisible, setVisibility] = useState<boolean>(!isGqlAPI)
    const [hasIntersected, setHasIntersected] = useState<boolean>(!isGqlAPI)

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
        if (!element || !isGqlAPI) {
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
    }, [isGqlAPI, reference])

    return { data, loading, error, isVisible, query }
}
