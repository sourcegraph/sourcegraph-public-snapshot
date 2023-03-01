import { createContext, FC, useCallback, useContext, useEffect, useMemo, useState } from 'react'

import { setCodeIntelSearchContext } from '@sourcegraph/shared/src/codeintel/searchContext'
import { getDefaultSearchContextSpec, isSearchContextSpecAvailable } from '@sourcegraph/shared/src/search'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { Subscription } from 'rxjs'
import { useLegacyPlatformContext } from './LegacyRouteContext'
import { parseSearchURL } from './search'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'

export const GLOBAL_SEARCH_CONTEXT_SPEC = 'global'

interface SearchContextType {
    setSelectedSearchContextSpec: (spec: string) => void
    /**
     * TODO: Move all the search context logic as close as possible to the components
     * that actually need it. Remove related `useState` from the `SourcegraphWebApp` component.
     */
    selectedSearchContextSpec: string | undefined
}
const SearchContext = createContext<SearchContextType>({
    setSelectedSearchContextSpec: () => undefined,
    selectedSearchContextSpec: undefined,
})

export function useSearchContext(): SearchContextType {
    return useContext(SearchContext)
}

export const SearchContextProvider: FC<React.PropsWithChildren<{ searchContextsEnabled: boolean }>> = props => {
    const [subscriptions] = useState(() => new Subscription())
    const platformContext = useLegacyPlatformContext()

    const [selectedSearchContextSpec, _setSelectedSearchContextSpec] = useState<string | undefined>()

    console.log({ selectedSearchContextSpec })

    // NOTE(2022-09-08) Inform the inlined code from sourcegraph/code-intel-extensions about the
    // change of search context. The old extension code previously accessed this information from
    // the 'sourcegraph' npm package, and updating the context like this was the simplest solution
    // to mirror the old behavior while deprecating extensions on a tight deadline. It would be nice
    // to properly pass around this via React state in the future.
    const setWorkspaceSearchContext = useCallback((spec: string | null): void => {
        setCodeIntelSearchContext(spec ?? undefined)
    }, [])

    const setSelectedSearchContextSpecWithNoChecks = useCallback(
        (spec: string): void => {
            console.log({ spec })
            _setSelectedSearchContextSpec(spec)
            setWorkspaceSearchContext(spec)
        },
        [setWorkspaceSearchContext]
    )
    const setSelectedSearchContextSpecToDefault = useCallback((): void => {
        if (!props.searchContextsEnabled) {
            return
        }
        subscriptions.add(
            getDefaultSearchContextSpec({ platformContext }).subscribe(spec => {
                console.log('getDefaultSearchContextSpec', { spec })
                // Fall back to global if no default is returned.
                setSelectedSearchContextSpecWithNoChecks(spec || GLOBAL_SEARCH_CONTEXT_SPEC)
            })
        )
    }, [platformContext, props.searchContextsEnabled, setSelectedSearchContextSpecWithNoChecks, subscriptions])

    const setSelectedSearchContextSpec = useCallback(
        (spec: string): void => {
            if (!props.searchContextsEnabled) {
                return
            }

            // The global search context is always available.
            if (spec === GLOBAL_SEARCH_CONTEXT_SPEC) {
                setSelectedSearchContextSpecWithNoChecks(spec)
            }

            // Check if the wanted search context is available.
            subscriptions.add(
                isSearchContextSpecAvailable({
                    spec,
                    platformContext,
                }).subscribe(isAvailable => {
                    if (isAvailable) {
                        setSelectedSearchContextSpecWithNoChecks(spec)
                    } else if (!selectedSearchContextSpec) {
                        // If the wanted search context is not available and
                        // there is no currently selected search context,
                        // set the current selection to the default search context.
                        // Otherwise, keep the current selection.
                        setSelectedSearchContextSpecToDefault()
                    }
                })
            )
        },
        [
            platformContext,
            props.searchContextsEnabled,
            selectedSearchContextSpec,
            setSelectedSearchContextSpecToDefault,
            setSelectedSearchContextSpecWithNoChecks,
            subscriptions,
        ]
    )

    useEffect(() => {
        const parsedSearchURL = parseSearchURL(window.location.search)
        const parsedSearchQuery = parsedSearchURL.query || ''

        if (parsedSearchQuery && !filterExists(parsedSearchQuery, FilterType.context)) {
            // If a context filter does not exist in the query, we have to switch the selected context
            // to global to match the UI with the backend semantics (if no context is specified in the query,
            // the query is run in global context).
            setSelectedSearchContextSpecWithNoChecks(GLOBAL_SEARCH_CONTEXT_SPEC)
        }
        if (!parsedSearchQuery) {
            // If no query is present (e.g. search page, settings page),
            // select the user's default search context.
            setSelectedSearchContextSpecToDefault()
        }

        setWorkspaceSearchContext(selectedSearchContextSpec ?? null)

        // We only ever want to run this hook once when the component mounts for
        // parity with the old behavior.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const context = useMemo(
        () => ({
            setSelectedSearchContextSpec,
            selectedSearchContextSpec,
        }),
        [setSelectedSearchContextSpec, selectedSearchContextSpec]
    )

    return <SearchContext.Provider value={context}>{props.children}</SearchContext.Provider>
}
