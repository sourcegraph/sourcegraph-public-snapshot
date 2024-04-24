import React, { createContext, useContext, useEffect, useState } from 'react'

import { useLocation } from 'react-router-dom'

import { useQuery, gql } from '@sourcegraph/http-client'

import {
    CodyIgnoreContentResult,
    CodyIgnoreContentVariables,
    ContextFiltersResult,
    ContextFiltersVariables,
} from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

import { isCodyEnabled } from './isCodyEnabled'

interface CodyIgnoreFns {
    isRepoIgnored(repoName: string): boolean
    isFileIgnored(repoName: string, filePath: string): boolean
}

export const alwaysTrue = (): true => true
export const alwaysFalse = (): false => false

const defaultCodyIgnoreFns = { isRepoIgnored: alwaysTrue, isFileIgnored: alwaysTrue }
const CodyIgnoreContext = createContext<CodyIgnoreFns>(defaultCodyIgnoreFns)

export const CodyIgnoreProvider: React.FC<React.PropsWithChildren<{ isSourcegraphDotCom: boolean }>> = ({
    isSourcegraphDotCom,
    children,
}) => (
    <CodyIgnoreContext.Provider
        value={
            isCodyEnabled()
                ? (isSourcegraphDotCom ? useCodyIgnoreFileFromRepo : useCodyContextFiltersFromSiteConfig)()
                : defaultCodyIgnoreFns
        }
    >
        {children}
    </CodyIgnoreContext.Provider>
)

export function useCodyIgnore(): CodyIgnoreFns {
    return useContext(CodyIgnoreContext)
}

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

const CODY_IGNORE_FILE_PATH = '.cody/ignore'

const useCodyIgnoreFileFromRepo = (): CodyIgnoreFns => {
    const location = useLocation()
    const { repoName, revision } = parseBrowserRepoURL(location.pathname + location.search + location.hash)
    const { data } = useQuery<CodyIgnoreContentResult, CodyIgnoreContentVariables>(CODY_IGNORE_CONTENT, {
        skip: !window.context?.experimentalFeatures.codyContextIgnore,
        variables: { repoName, repoRev: revision || '', filePath: CODY_IGNORE_FILE_PATH },
    })
    const [fns, setFns] = useState<CodyIgnoreFns>({
        isRepoIgnored: alwaysFalse,
        isFileIgnored: alwaysFalse,
    })

    const content = data?.repository?.commit?.blob?.content
    useEffect(() => {
        const createFilterFns = async (): Promise<void> => {
            if (content) {
                const ignore = (await import('ignore')).default
                setFns({ isRepoIgnored: alwaysFalse, isFileIgnored: ignore().add(content).ignores })
                return
            }
            setFns({ isRepoIgnored: alwaysFalse, isFileIgnored: alwaysFalse })
        }

        void createFilterFns()
    }, [content])

    return fns
}

const CODY_CONTEXT_FILTERS_QUERY = gql`
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

const useCodyContextFiltersFromSiteConfig = (): CodyIgnoreFns => {
    const { data, error, loading } = useQuery<ContextFiltersResult, ContextFiltersVariables>(
        CODY_CONTEXT_FILTERS_QUERY,
        {}
    )
    const [fns, setFns] = useState<CodyIgnoreFns>({
        isRepoIgnored: alwaysFalse,
        isFileIgnored: alwaysFalse,
    })

    useEffect(() => {
        const createFilterFns = async (): Promise<void> => {
            // Cody context filters are not available, ignore everything
            if (loading || error) {
                setFns({ isRepoIgnored: alwaysTrue, isFileIgnored: alwaysTrue })
                return
            }

            const filters = data?.site.codyContextFilters.raw as CodyContextFilters

            // Cody context filters are not defined, allow everything
            if (!filters) {
                setFns({ isRepoIgnored: alwaysFalse, isFileIgnored: alwaysFalse })
                return
            }

            setFns(await getFilterFnsFromCodyContextFilters(filters))
        }

        void createFilterFns()
    }, [loading, error, data])

    return fns
}

export async function getFilterFnsFromCodyContextFilters(filters: CodyContextFilters): Promise<CodyIgnoreFns> {
    const { RE2JS } = await import('re2js')
    let include: InstanceType<typeof RE2JS>[] = []
    let exclude: InstanceType<typeof RE2JS>[] = []
    try {
        include = filters.include?.map(({ repoNamePattern }) => RE2JS.compile(repoNamePattern)) || []
        exclude = filters.exclude?.map(({ repoNamePattern }) => RE2JS.compile(repoNamePattern)) || []
    } catch (error) {
        // eslint-disable-next-line no-console
        console.error('Failed to parse Cody context filters', error)
        // If we fail to parse the filters, ignore everything
        return { isRepoIgnored: alwaysTrue, isFileIgnored: alwaysTrue }
    }

    const isRepoIgnored = (repoName: string): boolean => {
        const isIncluded = !include.length || include.some(re => re.matches(repoName))
        const isExcluded = exclude.some(re => re.matches(repoName))
        return !isIncluded || isExcluded
    }

    // We don't support file-level ignore filters yet, so we just use the repo-level filters
    const isFileIgnored = (repoName: string, _: string): boolean => isRepoIgnored(repoName)
    return { isRepoIgnored, isFileIgnored }
}
