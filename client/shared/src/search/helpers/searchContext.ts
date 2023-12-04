import type { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'

import type { PlatformContext } from '../../platform/context'
import {
    type fetchSearchContexts,
    type getUserSearchContextNamespaces,
    type fetchSearchContext,
    type fetchSearchContextBySpec,
    type createSearchContext,
    type updateSearchContext,
    type deleteSearchContext,
    isSearchContextAvailable,
    fetchDefaultSearchContextSpec,
} from '../backend'

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
