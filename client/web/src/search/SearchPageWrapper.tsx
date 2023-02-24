import React from 'react'

import { useLocation } from 'react-router-dom-v5-compat'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LegacyLayoutRouteComponentProps } from '../routes'

import { parseSearchURLQuery } from '.'

const SearchPage = lazyComponent(() => import('./home/SearchPage'), 'SearchPage')
const StreamingSearchResults = lazyComponent(() => import('./results/StreamingSearchResults'), 'StreamingSearchResults')

/**
 * Renders the Search home page or Search results depending on whether a query
 * was submitted (present in the URL) or not.
 */
export const SearchPageWrapper: React.FunctionComponent<
    React.PropsWithChildren<LegacyLayoutRouteComponentProps>
> = props => {
    const location = useLocation()
    const hasSearchQuery = parseSearchURLQuery(location.search)

    return hasSearchQuery ? (
        <TraceSpanProvider name="StreamingSearchResults">
            <StreamingSearchResults {...props} />
        </TraceSpanProvider>
    ) : (
        <TraceSpanProvider name="SearchPage">
            <SearchPage {...props} />
        </TraceSpanProvider>
    )
}
