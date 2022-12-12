import { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import {
    fetchSearchContexts,
    getUserSearchContextNamespaces,
    fetchSearchContext,
    fetchSearchContextBySpec,
    createSearchContext,
    updateSearchContext,
    deleteSearchContext,
    isSearchContextAvailable,
    fetchDefaultSearchContextSpec,
} from './backend'
import { SearchPatternType } from './graphql-operations'
import { SearchMode } from './searchQueryState'

export * from './backend'
export * from './searchQueryState'
export * from './helpers'
export * from './graphql-operations'
export * from './helpers/queryExample'
export * from './integration/streaming-search-mocks'

export interface SearchPatternTypeProps {
    patternType: SearchPatternType
}

export interface SearchPatternTypeMutationProps {
    setPatternType: (patternType: SearchPatternType) => void
}

export interface CaseSensitivityProps {
    caseSensitive: boolean
    setCaseSensitivity: (caseSensitive: boolean) => void
}

export interface SearchModeProps {
    searchMode: SearchMode
    setSearchMode: (searchMode: SearchMode) => void
}

export interface SearchContextProps {
    searchContextsEnabled: boolean
    selectedSearchContextSpec?: string
    setSelectedSearchContextSpec: (spec: string) => void
    getUserSearchContextNamespaces: typeof getUserSearchContextNamespaces
    fetchSearchContexts: typeof fetchSearchContexts
    isSearchContextSpecAvailable: typeof isSearchContextSpecAvailable
    fetchSearchContext: typeof fetchSearchContext
    fetchSearchContextBySpec: typeof fetchSearchContextBySpec
    createSearchContext: typeof createSearchContext
    updateSearchContext: typeof updateSearchContext
    deleteSearchContext: typeof deleteSearchContext
}

export type SearchContextInputProps = Pick<
    SearchContextProps,
    | 'searchContextsEnabled'
    | 'selectedSearchContextSpec'
    | 'setSelectedSearchContextSpec'
    | 'fetchSearchContexts'
    | 'getUserSearchContextNamespaces'
>

export const isSearchContextSpecAvailable = memoizeObservable(
    ({ spec, platformContext }: { spec: string; platformContext: Pick<PlatformContext, 'requestGraphQL'> }) =>
        isSearchContextAvailable(spec, platformContext),
    ({ spec }) => spec
)

export const getAvailableSearchContextSpecOrFallback = memoizeObservable(
    ({
        spec,
        fallbackSpec,
        platformContext,
    }: {
        spec: string
        fallbackSpec: string
        platformContext: Pick<PlatformContext, 'requestGraphQL'>
    }) => isSearchContextAvailable(spec, platformContext).pipe(map(isAvailable => (isAvailable ? spec : fallbackSpec))),
    ({ spec, fallbackSpec }) => `${spec}:${fallbackSpec}`
)

export const getDefaultSearchContextSpec = memoizeObservable(
    ({ platformContext }: { platformContext: Pick<PlatformContext, 'requestGraphQL'> }): Observable<string> =>
        fetchDefaultSearchContextSpec(platformContext).pipe(
            map(spec => spec || ''),
            catchError(() => '')
        ),
    () => 'default'
)
