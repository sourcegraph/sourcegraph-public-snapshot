import H from 'history'
import { useCallback } from 'react'

/**
 * React component props for children of {@link WithQueryParameter}.
 */
export interface QueryParameterProps {
    /** The query. */
    query: string

    /** Called when the query changes. */
    onQueryChange: (query: string) => void

    /**
     * Called to obtain the URL for the query. Navigating to this URL is equivalent to calling
     * {@link QueryParameterProps#onQueryChange}.
     */
    locationWithQuery: (query: string) => H.LocationDescriptorObject
}

/**
 * A React hook for components that interact with the URL query parameter `q`.
 */
export const useQueryParameter = (
    { location, history }: { location: H.Location; history: H.History },
    defaultQuery = ''
): [QueryParameterProps['query'], QueryParameterProps['onQueryChange'], QueryParameterProps['locationWithQuery']] => {
    const q = new URLSearchParams(location.search).get('q')
    const query = q === null ? defaultQuery : q
    const locationWithQuery = useCallback(
        (query: string): H.LocationDescriptorObject => {
            const params = new URLSearchParams(location.search)
            params.set('q', query)
            return { ...location, hash: '', search: params.toString() }
        },
        [location]
    )
    const onQueryChange = useCallback(
        (query: string): void => {
            history.push(locationWithQuery(query))
        },
        [history, locationWithQuery]
    )
    return [query, onQueryChange, locationWithQuery]
}
