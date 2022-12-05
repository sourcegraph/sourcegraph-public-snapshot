import React, { useEffect, useMemo, useRef } from 'react'

import { Subject, Observable } from 'rxjs'
import { debounceTime, distinct, filter, map, mergeMap } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common/src'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client/src'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { RepoSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/wildcard/src'

import { FilesResult, FilesVariables } from '../graphql-operations'

interface FileResult {
    name: string
    path: string
}

interface TreeSearchResultsProps extends RepoSpec, ResolvedRevisionSpec, Pick<PlatformContext, 'requestGraphQL'> {
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

export const TreeSearchResults: React.FC<TreeSearchResultsProps> = props => {
    const results = useTreeSearchResults(props)

    return (
        <ul>
            {results?.map(file => (
                <TreeSearchResult key={file.path} file={file} />
            ))}
        </ul>
    )
}

const TreeSearchResult: React.FC<{ file: FileResult }> = ({ file }) => {
    const folderName = file.path.replace(new RegExp(`${file.name}$`), '')
    return (
        <li>
            <div>{file.name}</div>
            <div>{folderName}</div>
        </li>
    )
}
