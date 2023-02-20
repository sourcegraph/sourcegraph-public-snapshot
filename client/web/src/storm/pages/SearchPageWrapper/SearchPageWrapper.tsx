import { FC } from 'react'

import { useLocation } from 'react-router-dom'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { lazyComponent, lazyStormComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LegacyLayoutRouteComponentProps } from '../routes'

import { parseSearchURLQuery } from '../../../search'

const SearchPage = lazyStormComponent(() => import('../SearchPage/SearchPage'), 'SearchPage')
const StreamingSearchResults = lazyComponent(
    () => import('../../../search/results/StreamingSearchResults'),
    'StreamingSearchResults'
)

/**
 * Renders the Search home page or Search results depending on whether a query
 * was submitted (present in the URL) or not.
 */
export const SearchPageWrapper: FC<LegacyLayoutRouteComponentProps> = props => {
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
