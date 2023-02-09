import React from 'react'

import { useLocation } from 'react-router'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LayoutRouteComponentProps } from '../routes'

import { SearchPage } from './home/SearchPage'

import { parseSearchURLQuery } from '.'

const StreamingSearchResults = lazyComponent(() => import('./results/StreamingSearchResults'), 'StreamingSearchResults')

/**
 * Renders the Search home page or Search results depending on whether a query
 * was submitted (present in the URL) or not.
 */
export const SearchPageWrapper: React.FunctionComponent<React.PropsWithChildren<LayoutRouteComponentProps>> = props => {
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
