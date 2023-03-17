import { FC } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LegacyLayoutRouteContext } from '../LegacyRouteContext'

import { parseSearchURLQuery } from '.'

const SearchPage = lazyComponent(() => import('./home/SearchPage'), 'SearchPage')
const CodyHomepage = lazyComponent(() => import('../enterprise/cody/home/CodyHomepage'), 'CodyHomepage')
const StreamingSearchResults = lazyComponent(() => import('./results/StreamingSearchResults'), 'StreamingSearchResults')

const USE_CODY_SEARCH_PAGE = true

/**
 * Renders the Search home page or Search results depending on whether a query
 * was submitted (present in the URL) or not.
 */
export const SearchPageWrapper: FC<LegacyLayoutRouteContext> = props => {
    const location = useLocation()
    const navigate = useNavigate()
    const hasSearchQuery = parseSearchURLQuery(location.search)

    return hasSearchQuery ? (
        <TraceSpanProvider name="StreamingSearchResults">
            <StreamingSearchResults {...props} />
        </TraceSpanProvider>
    ) : (
        <TraceSpanProvider name="SearchPage">
            {USE_CODY_SEARCH_PAGE ? <CodyHomepage navigate={navigate} {...props} /> : <SearchPage {...props} />}
        </TraceSpanProvider>
    )
}
