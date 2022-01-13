import React from 'react'

import { LayoutRouteComponentProps } from '../routes'
import { useNavbarQueryState } from '../stores'
import { lazyComponent } from '../util/lazyComponent'

const SearchPage = lazyComponent(() => import('./home/SearchPage'), 'SearchPage')
const StreamingSearchResults = lazyComponent(() => import('./results/StreamingSearchResults'), 'StreamingSearchResults')

/**
 * Renders the Search home page or Search results depending on whether a query
 * was submitted (present in the URL) or not.
 */
export const SearchPageWrapper: React.FunctionComponent<LayoutRouteComponentProps<any>> = props => {
    const hasSearchQuery = useNavbarQueryState(state => state.searchQueryFromURL !== '')

    return hasSearchQuery ? <StreamingSearchResults {...props} /> : <SearchPage {...props} />
}
