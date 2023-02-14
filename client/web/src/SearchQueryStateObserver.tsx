import { FC, useEffect, useMemo } from 'react'

import { Location, useLocation } from 'react-router-dom'
import { Subject } from 'rxjs'
import { first } from 'rxjs/operators'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { isSearchContextSpecAvailable } from '@sourcegraph/shared/src/search'

import { getQueryStateFromLocation } from './search'
import { setQueryStateFromURL } from './stores/navbarSearchQueryState'

export const GLOBAL_SEARCH_CONTEXT_SPEC = 'global'

interface SearchQueryStateObserverProps {
    searchContextsEnabled: boolean
    platformContext: PlatformContext
    selectedSearchContextSpec?: string
    setSelectedSearchContextSpec: (spec: string) => void
}

// Update search query state whenever the URL changes
export const SearchQueryStateObserver: FC<SearchQueryStateObserverProps> = props => {
    const { searchContextsEnabled, platformContext, setSelectedSearchContextSpec, selectedSearchContextSpec } = props

    const location = useLocation()
    const locationSubject = useMemo(() => new Subject<Location>(), [])

    useEffect(() => locationSubject.next(location), [location, locationSubject])

    useEffect(() => {
        const subscription = getQueryStateFromLocation({
            location: locationSubject,
            showSearchContext: searchContextsEnabled,
            isSearchContextAvailable: (searchContext: string) =>
                searchContextsEnabled
                    ? isSearchContextSpecAvailable({ spec: searchContext, platformContext }).pipe(first()).toPromise()
                    : Promise.resolve(false),
        }).subscribe(parsedSearchURLAndContext => {
            if (parsedSearchURLAndContext.query) {
                // Only override filters and update query from URL if there
                // is a search query.
                if (
                    parsedSearchURLAndContext.searchContextSpec &&
                    parsedSearchURLAndContext.searchContextSpec !== selectedSearchContextSpec
                ) {
                    setSelectedSearchContextSpec(parsedSearchURLAndContext.searchContextSpec)
                } else if (!parsedSearchURLAndContext.searchContextSpec) {
                    // If no search context is present we have to fall back
                    // to the global search context to match the server
                    // behavior.
                    setSelectedSearchContextSpec(GLOBAL_SEARCH_CONTEXT_SPEC)
                }

                setQueryStateFromURL(parsedSearchURLAndContext, parsedSearchURLAndContext.processedQuery)
            }
        })

        return () => subscription.unsubscribe()
    }, [
        locationSubject,
        platformContext,
        searchContextsEnabled,
        selectedSearchContextSpec,
        setSelectedSearchContextSpec,
    ])

    return null
}
