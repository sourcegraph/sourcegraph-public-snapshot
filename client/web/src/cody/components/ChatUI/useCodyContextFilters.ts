import { useEffect, useState } from 'react'

import { useLocation } from 'react-router-dom'

import { useQuery, gql } from '@sourcegraph/http-client'

import {
    CodyIgnoreContentResult,
    CodyIgnoreContentVariables,
    ContextFiltersResult,
    ContextFiltersVariables,
} from '../../../graphql-operations'
import { parseBrowserRepoURL } from '../../../util/url'

export interface CodyContextFiltersFns {
    isRepoIncluded(repoName: string): boolean
    isFileIncluded(repoName: string, filePath: string): boolean
}
type UseCodyContextFilters = () => CodyContextFiltersFns

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

const useDotcomContextFilters: UseCodyContextFilters = () => {
    const location = useLocation()
    const { repoName, revision } = parseBrowserRepoURL(location.pathname + location.search + location.hash)
    const { data } = useQuery<CodyIgnoreContentResult, CodyIgnoreContentVariables>(CODY_IGNORE_CONTENT, {
        skip: !window.context?.experimentalFeatures.codyContextIgnore,
        variables: { repoName, repoRev: revision || '', filePath: CODY_IGNORE_PATH },
    })
    const [filterFns, setFilterFns] = useState<CodyContextFiltersFns>({
        isRepoIncluded: () => true,
        isFileIncluded: () => true,
    })

    const content = data?.repository?.commit?.blob?.content
    useEffect(() => {
        const createFilterFns = async (): Promise<void> => {
            if (content) {
                const ignore = (await import('ignore')).default
                setFilterFns({ isRepoIncluded: () => true, isFileIncluded: ignore().add(content).ignores })
                return
            }
            setFilterFns({ isRepoIncluded: () => true, isFileIncluded: () => true })
        }

        void createFilterFns()
    }, [content])

    return filterFns
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

interface CodyContextFilters {
    include?: CodyContextFilterItem[]
    exclude?: CodyContextFilterItem[]
}

interface CodyContextFilterItem {
    repoNamePattern: string
}

const useEnterpriseCodyContextFilters: UseCodyContextFilters = () => {
    const { data, error, loading } = useQuery<ContextFiltersResult, ContextFiltersVariables>(CONTEXT_FILTERS_QUERY, {})
    const [filterFns, setFilterFns] = useState<CodyContextFiltersFns>({
        isRepoIncluded: () => false,
        isFileIncluded: () => false,
    })

    useEffect(() => {
        const createFilterFns = async (): Promise<void> => {
            if (loading || error) {
                setFilterFns({ isRepoIncluded: () => false, isFileIncluded: () => false })
                return
            }

            const filters = data?.site.codyContextFilters.raw as CodyContextFilters
            if (!filters) {
                setFilterFns({ isRepoIncluded: () => true, isFileIncluded: () => true })
            }

            const { RE2 } = await import('re2-wasm')
            let include: InstanceType<typeof RE2>[] = []
            let exclude: InstanceType<typeof RE2>[] = []
            try {
                include = filters.include?.map(({ repoNamePattern }) => new RE2(repoNamePattern, 'u')) || []
                exclude = filters.exclude?.map(({ repoNamePattern }) => new RE2(repoNamePattern, 'u')) || []
            } catch (error) {
                // eslint-disable-next-line no-console
                console.error('Failed to parse Cody context filters', error)
                setFilterFns({ isRepoIncluded: () => false, isFileIncluded: () => false })
                return
            }

            const isRepoIncluded = (repoName: string): boolean => {
                let isIncluded = true
                for (const re of include) {
                    isIncluded = re.test(repoName)
                    if (isIncluded) {
                        break
                    }
                }
                for (const re of exclude) {
                    if (re.test(repoName)) {
                        isIncluded = false
                        break
                    }
                }
                return isIncluded
            }
            const isFileIncluded = (repoName: string, _: string): boolean => isRepoIncluded(repoName)
            setFilterFns({ isRepoIncluded, isFileIncluded })
        }

        void createFilterFns()
    }, [loading, error, data])

    return filterFns
}

export const getUseCodyContextFiltersHook = (isSourcegraphDotCom: boolean): UseCodyContextFilters =>
    isSourcegraphDotCom ? useDotcomContextFilters : useEnterpriseCodyContextFilters
