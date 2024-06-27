import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'

import { useApolloClient, type ApolloClient, type ApolloQueryResult } from '@apollo/client'

import { getDocumentNode, gql, useQuery } from '@sourcegraph/http-client'

import type {
    CodyIgnoreContentResult,
    CodyIgnoreContentVariables,
    ContextFiltersResult,
    ContextFiltersVariables,
} from '../graphql-operations'

interface CodyIgnoreFns {
    isRepoIgnored(repoName: string): boolean
    isFileIgnored(repoName: string, filePath: string): boolean
}

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

export function useCodyIgnore(): CodyIgnoreFns {
    return useContext(CodyIgnoreContext)
}

/**
 * CodyIgnoreProvider provides {@link CodyIgnoreFns} based on the {@link isSourcegraphDotCom} prop.
 *
 * If Cody is enabled, the {@link CodyIgnoreFns} are defined based on:
 * - {@link CODY_IGNORE_FILE_PATH} file content for dotcom users;
 * - {@link CodyContextFilters} from the site config for enterprise users.
 *
 * If Cody is not enabled, {@link defaultCodyIgnoreFns} are used.
 */
export const CodyIgnoreProvider: React.FC<React.PropsWithChildren<{ isSourcegraphDotCom: boolean }>> = ({
    isSourcegraphDotCom,
    children,
}) => {
    // Cody is not enabled, return default ignore fns.
    if (!window.context?.codyEnabledForCurrentUser) {
        return <CodyIgnoreContext.Provider value={defaultCodyIgnoreFns}>{children}</CodyIgnoreContext.Provider>
    }

    if (isSourcegraphDotCom) {
        // Cody Ignore is an experimental feature on dotcom. If the feature is not enabled, return default ignore fns.
        if (!window.context?.experimentalFeatures.codyContextIgnore) {
            return <CodyIgnoreContext.Provider value={defaultCodyIgnoreFns}>{children}</CodyIgnoreContext.Provider>
        }

        return <DotcomProvider>{children}</DotcomProvider>
    }

    return <EnterpriseProvider>{children}</EnterpriseProvider>
}

/**
 * DotcomProvider provides {@link CodyIgnoreFns} as {@link CodyIgnoreContext} value
 * based on the {@link CODY_IGNORE_FILE_PATH} content for the given repo.
 *
 * Rules defined in {@link CODY_IGNORE_FILE_PATH} apply only to a given repo, thus:
 * - {@link CodyIgnoreFns.isRepoIgnored} always returns `false`
 * - {@link CodyIgnoreFns.isFileIgnored} returns whether the file path matches ignore rules if
 * {@link CODY_IGNORE_FILE_PATH} exists in the repo, and `false` if it doesn't exist.
 */
const DotcomProvider: React.FC<React.PropsWithChildren<{}>> = ({ children }) => {
    const client = useApolloClient()
    const [isFileIgnoredByRepo, setIsFileIgnoredByRepo] = useState<Record<string, CodyIgnoreFns['isFileIgnored']>>({})

    const fetchCodyIgnoreContent = useMemo(() => createCodyIgnoreContentFetcher(client), [client])

    const isFileIgnored: CodyIgnoreFns['isFileIgnored'] = useCallback(
        (repoName, filePath) => {
            // If ignore fn is defined for the repo, use it.
            const fn = isFileIgnoredByRepo[repoName]
            if (fn) {
                return fn(repoName, filePath)
            }

            // If ignore fn is not defined for the repo, fetch the ignore file content and update the state.
            void fetchCodyIgnoreContent(repoName).then(fn =>
                setIsFileIgnoredByRepo(state => ({ ...state, [repoName]: fn }))
            )

            // Ignore file as we don't have the ignore file content yet.
            return true
        },
        [isFileIgnoredByRepo, setIsFileIgnoredByRepo, fetchCodyIgnoreContent]
    )

    const fns = useMemo(() => ({ isRepoIgnored: alwaysFalse, isFileIgnored }), [isFileIgnored])

    return <CodyIgnoreContext.Provider value={fns}>{children}</CodyIgnoreContext.Provider>
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
 * createCodyIgnoreContentFetcher creates a function that fetches {@link CODY_IGNORE_FILE_PATH} content for a given repo.
 *
 * If an error occurs when fetching the content, the function returns {@link alwaysTrue} function.
 *
 * If the content is fetched successfully, the function returns:
 * - if the ignore file exists - {@link CodyIgnoreFns.isFileIgnored} function based on the ignore rules from the content
 * - if the ignore file doesn't exist - {@link alwaysFalse} function.
 *
 * Function caches promises to avoid fetching the content for the same repo multiple times.
 */
function createCodyIgnoreContentFetcher(
    client: ApolloClient<unknown>
): (repoName: string) => Promise<CodyIgnoreFns['isFileIgnored']> {
    const cache = new Map<string, Promise<ApolloQueryResult<CodyIgnoreContentResult>>>()

    return async (repoName: string) => {
        let promise = cache.get(repoName)
        if (!promise) {
            promise = client.query<CodyIgnoreContentResult, CodyIgnoreContentVariables>({
                query: getDocumentNode(CODY_IGNORE_CONTENT),
                variables: { repoName, repoRev: 'HEAD', filePath: CODY_IGNORE_FILE_PATH },
            })
            cache.set(repoName, promise)
        }

        const { error, data } = await promise

        if (error) {
            // Error when fetching ignore file, ignore everything.
            return alwaysTrue
        }

        const content = data?.repository?.commit?.blob?.content
        if (content) {
            const ignore = (await import('ignore')).default
            return (_repoName: string, filePath: string) => ignore().add(content).ignores(filePath)
        }

        // Ignore file doesn't exist for a given repo, allow everything.
        return alwaysFalse
    }
}

interface CodyContextFilters {
    include?: CodyContextFilterItem[]
    exclude?: CodyContextFilterItem[]
}

interface CodyContextFilterItem {
    // Regex following RE2 syntax
    repoNamePattern: string
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

/**
 * EnterpriseProvider fetches {@link CodyContextFilters} from the site config and provides {@link CodyIgnoreFns} as
 * {@link CodyIgnoreContext} value based on the filters from the site config:
 * - If the query hasn't yet started, filters are {@link CodyIgnoreFns} are set to ignore everything.
 * - If the site config query is loading or returned an error, filters are {@link CodyIgnoreFns} are set to ignore everything.
 * - If {@link CodyContextFilters} are not defined in the site config, {@link CodyIgnoreFns} are set to allow everything.
 * - If {@link CodyContextFilters} are defined, {@link CodyIgnoreFns} are set based on the filters value.
 */
const EnterpriseProvider: React.FC<React.PropsWithChildren<{}>> = ({ children }) => {
    const { data, error, loading } = useQuery<ContextFiltersResult, ContextFiltersVariables>(
        CODY_CONTEXT_FILTERS_QUERY,
        { fetchPolicy: 'cache-and-network' }
    )
    const [fns, setFns] = useState<CodyIgnoreFns>({ isRepoIgnored: alwaysTrue, isFileIgnored: alwaysTrue })

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

    return <CodyIgnoreContext.Provider value={fns}>{children}</CodyIgnoreContext.Provider>
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
