import { useCallback, useEffect, useState } from 'react'

import type { Ignore } from 'ignore'
import type { RE2 } from 're2-wasm'
import { useLocation } from 'react-router-dom'

import { useQuery, gql } from '@sourcegraph/http-client'

import {
    CodyIgnoreContentResult,
    CodyIgnoreContentVariables,
    ContextFiltersResult,
    ContextFiltersVariables,
} from '../../../graphql-operations'
import { parseBrowserRepoURL } from '../../../util/url'

type FilterFunc = (path: string) => boolean
type FilterHook = () => FilterFunc

const CODY_IGNORE_CONTENT = gql`
    query CodyIgnoreContent($repoName: String!, $repoRev: String!, $filePath: String!) {
        repository(name: $repoName) {
            commit(rev: $repoRev) {
                blob(path: $filePath) {
                    content
                }
            }
        }
    }
`

const CODY_IGNORE_PATH = '.cody/ignore'

const useDotcomContextFilter: FilterHook = () => {
    const location = useLocation()
    const { repoName, revision } = parseBrowserRepoURL(location.pathname + location.search + location.hash)
    const { data } = useQuery<CodyIgnoreContentResult, CodyIgnoreContentVariables>(CODY_IGNORE_CONTENT, {
        skip: !window.context?.experimentalFeatures.codyContextIgnore,
        variables: { repoName, repoRev: revision || '', filePath: CODY_IGNORE_PATH },
    })
    const [ignoreManager, setIgnoreManager] = useState<Ignore>()

    const content = data?.repository?.commit?.blob?.content
    useEffect(() => {
        const loadIgnore = async (): Promise<void> => {
            if (content) {
                const ignore = (await import('ignore')).default
                setIgnoreManager(ignore().add(content))
            }
        }

        void loadIgnore()
    }, [content])

    const filter = useCallback(
        (path: string): boolean => {
            if (ignoreManager) {
                return ignoreManager.ignores(path)
            }
            return false
        },
        [ignoreManager]
    )

    return filter
}

export const CONTEXT_FILTERS_QUERY = gql`
    query ContextFilters {
        site {
            codyContextFilters(version: V1) {
                raw
            }
        }
    }
`

interface ContextFilters {
    include?: CodyContextFilterItem[]
    exclude?: CodyContextFilterItem[]
}

interface CodyContextFilterItem {
    repoNamePattern: string
}

// const parseCodyContextFilters = (filters: ContextFilters)

const useEnterpriseCodyContextFilter: FilterHook = () => {
    // const location = useLocation()
    // const { repoName } = parseBrowserRepoURL(location.pathname + location.search + location.hash)
    const [createRE2, setCreateRE2] = useState<(...args: ConstructorParameters<typeof RE2>) => RE2>()
    const { data, error, loading } = useQuery<ContextFiltersResult, ContextFiltersVariables>(CONTEXT_FILTERS_QUERY, {})

    useEffect(() => {
        const loadRE2 = async (): Promise<void> => {
            if (data?.site.codyContextFilters.raw) {
                const { RE2 } = await import('re2-wasm')
                setCreateRE2((...args: ConstructorParameters<typeof RE2>) => new RE2(...args))
            }

            void loadRE2()
        }
    }, [data])

    const filter: FilterFunc = useCallback(
        path => {
            if (loading || error) {
                return false
            }
            const filters = data?.site.codyContextFilters.raw
            if (!filters) {
                return true
            }
            if (!createRE2) {
                return false
            }
            try {
                const { include, exclude } = filters as ContextFilters
            } catch (error_) {
                // eslint-disable-next-line no-console
                console.error('Error parsing Cody context filters:', error_)
                return false
            }

            // TODO: parse filters
            return true
        },
        [loading, error, data, createRE2]
    )

    return filter
}

export const getCodyContextFilterHook = (isSourcegraphDotCom: boolean): FilterHook =>
    isSourcegraphDotCom ? useDotcomContextFilter : useEnterpriseCodyContextFilter
