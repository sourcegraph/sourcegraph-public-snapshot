import React, { useEffect, useMemo, useRef } from 'react'

import { Subject, Observable } from 'rxjs'
import { debounceTime, distinct, filter, map, mergeMap } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common/src'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client/src'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { RepoSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Alert, Code, Text, Link, useObservable, LoadingSpinner } from '@sourcegraph/wildcard/src'

import { useShowMorePagination } from '../components/FilteredConnection/hooks/useShowMorePagination'
import { FilesResult, FilesVariables, SearchPatternType } from '../graphql-operations'
import { SearchStreamingProps } from '../search'
import { useCachedSearchResults } from '../search/results/SearchResultsCacheProvider'
import { LATEST_VERSION, SearchMatch } from '@sourcegraph/shared/out/src/search/stream'
import { ExtensionsControllerProps } from '@sourcegraph/shared/out/src/extensions/controller'
import { TelemetryProps } from '@sourcegraph/shared/out/src/telemetry/telemetryService'
import { VirtualList } from '@sourcegraph/shared/out/src/components/VirtualList'
import classNames from 'classnames'
import styles from '@sourcegraph/search-ui/src/results/StreamingSearchResultsList.module.scss'
import { escapeRegExp } from 'lodash'
// import { Link } from 'react-router-dom'

import treeStyles from './Tree.module.scss'

interface FileResult {
    name: string
    path: string
}

interface TreeSearchResultsProps
    extends RepoSpec,
        ResolvedRevisionSpec,
        Pick<PlatformContext, 'requestGraphQL'>,
        SearchStreamingProps,
        ExtensionsControllerProps,
        TelemetryProps {
    searchTerm: string
}

const FILES_QUERY = gql`
    query Files($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                results {
                    ... on FileMatch {
                        __typename
                        file {
                            name
                            path
                        }
                    }
                }
            }
        }
    }
`

const RESULTS_COUNT = 50

const fetchFiles = memoizeObservable(
    ({
        requestGraphQL,
        ...args
    }: Pick<PlatformContext, 'requestGraphQL'> & FilesVariables): Observable<FileResult[] | undefined> =>
        requestGraphQL<FilesResult, FilesVariables>({
            request: FILES_QUERY,
            variables: args,
            mightContainPrivateInfo: true,
        }).pipe(
            map(dataOrThrowErrors),
            map(({ search }) =>
                search?.results.results?.reduce((acc, result) => {
                    if (result.__typename === 'FileMatch') {
                        acc.push(result.file)
                    }
                    return acc
                }, [] as FileResult[])
            )
        ),
    ({ query }) => query
)

const useTreeSearchResults = ({
    searchTerm,
    repoName,
    commitID,
    requestGraphQL,
}: TreeSearchResultsProps): FileResult[] | undefined => {
    const searchTermRef = useRef<Subject<string>>(new Subject())

    useEffect(() => {
        searchTermRef.current.next(searchTerm)
    }, [searchTerm])

    return useObservable(
        useMemo(
            () =>
                searchTermRef.current.pipe(
                    debounceTime(500),
                    filter(Boolean),
                    distinct(),
                    mergeMap(term =>
                        fetchFiles({
                            requestGraphQL,
                            query: `repo:${repoName} revision:${commitID} type:path count:${RESULTS_COUNT} ${term}`,
                        })
                    )
                ),
            [requestGraphQL, repoName, commitID]
        )
    )
}

const useDebouncedSearchTerm = (searchTerm: string): string | undefined => {
    const searchTermRef = useRef<Subject<string>>(new Subject())

    useEffect(() => {
        searchTermRef.current.next(searchTerm)
    }, [searchTerm])

    return useObservable(
        useMemo(
            () =>
                searchTermRef.current.pipe(
                    debounceTime(500),
                    filter(value => !!value),
                    distinct()
                ),
            []
        )
    )
}

const TreeSearchResultList: React.FC<TreeSearchResultsProps> = props => {
    const query = `repo:${props.repoName} revision:${props.commitID} type:path count:${RESULTS_COUNT} ${props.searchTerm}`
    const trace = useMemo(() => new URLSearchParams(location.search).get('trace') ?? undefined, [])
    const options = useMemo(
        () => ({ version: LATEST_VERSION, patternType: SearchPatternType.regexp, caseSensitive: false, trace }),
        [trace]
    )
    const results = useCachedSearchResults(
        props.streamSearch,
        query,
        options,
        props.extensionsController !== null && window.context.enableLegacyExtensions
            ? props.extensionsController.extHostAPI
            : null,
        props.telemetryService
    )

    console.log(results)

    const filePaths = results?.results.reduce((acc, result) => {
        if (result.type === 'path') {
            acc.push(result.path)
        }
        return acc
    }, [] as string[])

    return (
        <>
            {filePaths ? (
                <ul className="list-unstyled">
                    {filePaths.map(path => (
                        <TreeSearchResult key={path} path={path} searchTerm={props.searchTerm} />
                    ))}
                </ul>
            ) : null}
            {(!results || results?.state === 'loading') && (
                <div className="text-center my-4" data-testid="loading-container">
                    <LoadingSpinner />
                </div>
            )}
            {results?.state === 'complete' &&
                results.progress.skipped.some(skipped => skipped.reason.includes('-limit')) && (
                    <Alert className="text-wrap" variant="info">
                        <Text className="m-0 text-center">
                            <strong>Result limit hit.</strong>
                        </Text>
                        <Text className="m-0">
                            Go to <Link to="/search">search results page</Link> and modify your search with{' '}
                            <Code>count:</Code> to return additional items.
                        </Text>
                    </Alert>
                )}
        </>
    )
}

export const TreeSearchResults: React.FC<TreeSearchResultsProps> = props => {
    const debouncedSearchTerm = useDebouncedSearchTerm(props.searchTerm)

    if (!debouncedSearchTerm) {
        return null
    }

    return <TreeSearchResultList {...props} searchTerm={debouncedSearchTerm} />
}

const toHighlightedElements = (items: string[], regex: RegExp): React.ReactNode[] =>
    items
        .filter(Boolean)
        .map((item, index) => (regex.test(item) ? <mark key={index}>{item}</mark> : <span key={index}>{item}</span>))

const TreeSearchResult: React.FC<{ path: string; searchTerm: string }> = ({ path, searchTerm }) => {
    const regex = new RegExp(`(${escapeRegExp(searchTerm)})`, 'i')
    const pathParts = path.split('/')
    const fileName = pathParts[pathParts.length - 1]
    const dirName = pathParts.slice(0, -1).join('/')

    const fileNameParts = []
    const dirNameParts = []

    const fileNameMatches = fileName.split(regex)
    if (fileNameMatches.length > 1) {
        fileNameParts.push(...toHighlightedElements(fileNameMatches, regex))
        dirNameParts.push(dirName)
    } else {
        fileNameParts.push(fileName)
    }

    if (dirNameParts.length === 0) {
        const dirNameMatches = dirName.split(regex)
        dirNameParts.push(...toHighlightedElements(dirNameMatches, regex))
    }

    return (
        <li className={classNames(treeStyles.row, 'px-1 border-bottom')}>
            <Link to={path} className="d-flex flex-column justify-space-between">
                <span>{fileNameParts}</span>
                <span>{dirNameParts}</span>
            </Link>
        </li>
    )
}
