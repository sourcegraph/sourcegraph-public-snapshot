import React from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LayoutRouteComponentProps } from '../routes'
import { useExperimentalFeatures, useNavbarQueryState } from '../stores'

import { ComputeSearchResults } from './results/ComputeSearchResults'

const SearchPage = lazyComponent(() => import('./home/SearchPage'), 'SearchPage')
const StreamingSearchResults = lazyComponent(() => import('./results/StreamingSearchResults'), 'StreamingSearchResults')

/**
 * Renders the Search home page or Search results depending on whether a query
 * was submitted (present in the URL) or not.
 */
export const SearchPageWrapper: React.FunctionComponent<
    React.PropsWithChildren<LayoutRouteComponentProps<any>>
> = props => {
    const searchQuery = useNavbarQueryState(state => state.searchQueryFromURL)
    const isComputeEnabled = useExperimentalFeatures(state => state.showComputeComponent)

    const hasSearchQuery = searchQuery !== ''
    const showComputeResults = isComputeEnabled && hasSearchQuery && searchQuery.includes('content:output(') // Naive check

    return hasSearchQuery ? (
        showComputeResults ? (
            <ComputeSearchResults {...props} />
        ) : (
            <StreamingSearchResults {...props} />
        )
    ) : (
        <SearchPage {...props} />
    )
}
