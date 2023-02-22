import { FC, useEffect, useRef, useState } from 'react'

import { Location, useLocation } from 'react-router-dom'
import { BehaviorSubject } from 'rxjs'
import { first } from 'rxjs/operators'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { isSearchContextSpecAvailable } from '@sourcegraph/shared/src/search'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'

import { getQueryStateFromLocation } from './search'
import { useExperimentalFeatures } from './stores'
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

    const selectedSearchContextSpecRef = useRef(selectedSearchContextSpec)
    selectedSearchContextSpecRef.current = selectedSearchContextSpec

    const { searchQueryInput, isInitialized } = useExperimentalFeatures()

    // This ensures that the query stays unmodified until we know
    // whether the feature flag is set or not.
    const enableExperimentalSearchInput = isInitialized ? searchQueryInput === 'experimental' : true
    const enableExperimentalSearchInputRef = useRef(enableExperimentalSearchInput)
    enableExperimentalSearchInputRef.current = enableExperimentalSearchInput

    // Create `locationSubject` once on mount. New values are provided in the `useEffect` hook.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const [locationSubject] = useState(() => new BehaviorSubject<Location>(location))

    useEffect(() => {
        locationSubject.next(location)
    }, [location, locationSubject])

    useEffect(() => {
        const subscription = getQueryStateFromLocation({
            location: locationSubject,
            isSearchContextAvailable: (searchContext: string) =>
                searchContextsEnabled
                    ? isSearchContextSpecAvailable({
                          spec: searchContext,
                          platformContext,
                      })
                          .pipe(first())
                          .toPromise()
                    : Promise.resolve(false),
        }).subscribe(parsedSearchURLAndContext => {
            if (parsedSearchURLAndContext.query) {
                // Only override filters and update query from URL if there
                // is a search query.
                if (!parsedSearchURLAndContext.searchContextSpec) {
                    // If no search context is present we have to fall back
                    // to the global search context to match the server
                    // behavior.
                    setSelectedSearchContextSpec(GLOBAL_SEARCH_CONTEXT_SPEC)
                } else if (parsedSearchURLAndContext.searchContextSpec.spec !== selectedSearchContextSpecRef.current) {
                    setSelectedSearchContextSpec(parsedSearchURLAndContext.searchContextSpec.spec)
                }

                const processedQuery =
                    !enableExperimentalSearchInputRef.current &&
                    parsedSearchURLAndContext.searchContextSpec &&
                    searchContextsEnabled
                        ? omitFilter(
                              parsedSearchURLAndContext.query,
                              parsedSearchURLAndContext.searchContextSpec.filter
                          )
                        : parsedSearchURLAndContext.query

                setQueryStateFromURL(parsedSearchURLAndContext, processedQuery)
            }
        })

        return () => subscription.unsubscribe()
    }, [
        locationSubject,
        platformContext,
        searchContextsEnabled,
        selectedSearchContextSpecRef,
        enableExperimentalSearchInputRef,
        setSelectedSearchContextSpec,
    ])

    return null
}
