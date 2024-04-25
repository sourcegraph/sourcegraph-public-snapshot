import React, { createContext, useCallback, useContext, useEffect, useState } from 'react'

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

type CodyIgnoreHook = (isRepositoryRelatedPage?: boolean) => CodyIgnoreFns

// alwaysTrue is exported only for testing purposes.
export const alwaysTrue = (): true => true
// alwaysFalse is exported only for testing purposes.
export const alwaysFalse = (): false => false

/**
 * defaultCodyIgnoreFns is used to set the default `CodyIgnoreContext` value.
 *
 * Every repo and file is allowed by default.
 */
const defaultCodyIgnoreFns = { isRepoIgnored: alwaysFalse, isFileIgnored: alwaysFalse }
const CodyIgnoreContext = createContext<CodyIgnoreFns>(defaultCodyIgnoreFns)

/**
 * CodyIgnoreProvider provides {@link CodyIgnoreFns} based on the {@link isSourcegraphDotCom} prop.
 *
 * If Cody is enabled, the {@link CodyIgnoreFns} are defined based on:
 * - {@link CODY_IGNORE_FILE_PATH} file content for dotcom users;
 * - {@link CodyContextFilters} from the site config for enterprise users.
 *
 * If Cody is not enabled, {@link defaultCodyIgnoreFns} are used.
 */
export const CodyIgnoreProvider: React.FC<
    React.PropsWithChildren<{ isSourcegraphDotCom: boolean; isRepositoryRelatedPage?: boolean }>
> = ({ isSourcegraphDotCom, isRepositoryRelatedPage, children }) => {
    const getCodyIgnoreFns = useCallback(() => {
        if (!isCodyEnabled() || (isSourcegraphDotCom && !window.context?.experimentalFeatures.codyContextIgnore)) {
            return () => defaultCodyIgnoreFns
        }
        return isSourcegraphDotCom ? useCodyIgnoreFileFromRepo : useCodyContextFiltersFromSiteConfig
    }, [isSourcegraphDotCom])
    return (
        <CodyIgnoreContext.Provider value={getCodyIgnoreFns()(isRepositoryRelatedPage)}>
            {children}
        </CodyIgnoreContext.Provider>
    )
}

export function useCodyIgnore(): CodyIgnoreFns {
    return useContext(CodyIgnoreContext)
}

const CODY_IGNORE_CONTENT = gql`
    query CodyIgnoreContent($repoName: String!, $repoRev: String!, $filePath: String!) {
        repository(name: $repoName) {
            id
            commit(rev: $repoRev) {
                id
                blob(path: $filePath) {
                    content
                }
            }
        }
    }
`

const CODY_IGNORE_FILE_PATH = '.cody/ignore'

/**
 * useCodyIgnoreFileFromRepo is a custom hook that fetches the current repository {@link CODY_IGNORE_FILE_PATH} content,
 * parses it following `gitignore` spec, and returns {@link CodyIgnoreFns} based on the file content.
 *
 * By design only file-level ignores are supported as the ignore rules defined in {@link CODY_IGNORE_FILE_PATH} apply
 * only to the current repository, thus:
 * - {@link CodyIgnoreFns.isRepoIgnored} always returns `false`
 * - {@link CodyIgnoreFns.isFileIgnored} returns whether the file path matches ignore rules if {@link CODY_IGNORE_FILE_PATH}
 * exists in the repository, and `false` if it doesn't exist.
 */
const useCodyIgnoreFileFromRepo: CodyIgnoreHook = isRepositoryRelatedPage => {
    const location = useLocation()
    // If the current page is repository-related, we can safely parse the repo name from the URL.
    const repoName = isRepositoryRelatedPage
        ? parseBrowserRepoURL(location.pathname + location.search + location.hash).repoName
        : ''
    const { data } = useQuery<CodyIgnoreContentResult, CodyIgnoreContentVariables>(CODY_IGNORE_CONTENT, {
        skip: !repoName,
        variables: { repoName, repoRev: 'HEAD', filePath: CODY_IGNORE_FILE_PATH },
    })
    const [fns, setFns] = useState<CodyIgnoreFns>({
        isRepoIgnored: alwaysFalse,
        isFileIgnored: alwaysFalse,
    })

    const content = data?.repository?.commit?.blob?.content
    useEffect(() => {
        // To dynamically import ignore parsing library, we need to call this function in `useEffect`.
        void (async () => {
            if (content) {
                const ignore = (await import('ignore')).default
                setFns({
                    isRepoIgnored: alwaysFalse,
                    isFileIgnored: (_repoName: string, filePath) => ignore().add(content).ignores(filePath),
                })
                return
            }

            setFns({ isRepoIgnored: alwaysFalse, isFileIgnored: alwaysFalse })
        })()
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
    // Regex following RE2 syntax
    repoNamePattern: string
}

/**
 * useCodyContextFiltersFromSiteConfig is a custom hook that fetches {@link CodyContextFilters} from the site config
 * and returns {@link CodyIgnoreFns} based on the filters value.
 *
 * If the site config query is loading or returned an error, filters are {@link CodyIgnoreFns} are set to ignore everything.
 *
 * If {@link CodyContextFilters} are not defined in the site config, {@link CodyIgnoreFns} are set to allow everything.
 *
 * If {@link CodyContextFilters} are defined, {@link CodyIgnoreFns} are set based on the filters value.
 *
 */
const useCodyContextFiltersFromSiteConfig: CodyIgnoreHook = () => {
    const { data, error, loading } = useQuery<ContextFiltersResult, ContextFiltersVariables>(
        CODY_CONTEXT_FILTERS_QUERY,
        {}
    )
    const [fns, setFns] = useState<CodyIgnoreFns>({
        isRepoIgnored: alwaysFalse,
        isFileIgnored: alwaysFalse,
    })

    useEffect(() => {
        // To dynamically import RE2 regex parsing library, we need to call this function in `useEffect`.
        void (async () => {
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
        })()
    }, [loading, error, data])

    return fns
}

/**
 * getFilterFnsFromCodyContextFilters imports RE2 regexes parsing library and returns {@link CodyIgnoreFns}.
 *
 * `isRepoIgnored` function returns true if repo doesn't match any of the `include` or matches any of `exclude` repo name patterns.
 *
 * Currently, only repo-level ignore filters are supported. Thus, `isFileIgnored` function returns the same result as `isRepoIgnored`.
 * If repo is allowed, every file in it is allowed too.
 *
 * If filters include repo name patterns that are not valid regexes, both `isRepoIgnored` and `isFileIgnored` functions return false.
 *
 * This function is exported only for testing purposes.
 */
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
    const isFileIgnored = (repoName: string, _filePath: string): boolean => isRepoIgnored(repoName)
    return { isRepoIgnored, isFileIgnored }
}
