import { useEffect, useState } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import {
    diffStatFieldsFragment,
    fileDiffFieldsFragment,
    fileDiffHunkRangeFieldsFragment,
} from '../../../../repo/compare/RepositoryCompareDiffPage'
import { gitRevisionRangeFieldsFragment } from '../../../../repo/compare/RepositoryCompareOverviewPage'

const LOADING: 'loading' = 'loading'

export function useThreadFileDiffs(
    thread: Pick<GQL.IThread, 'id'>
): typeof LOADING | GQL.IRepositoryComparison | ErrorLike {
    const [result, setResult] = useState<typeof LOADING | GQL.IRepositoryComparison | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryThreadFileDiffs(thread).subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [thread])
    return result
}

function queryThreadFileDiffs(thread: Pick<GQL.IThread, 'id'>): Observable<GQL.IRepositoryComparison> {
    return queryGraphQL(
        gql`
            query ThreadFileDiffs($thread: ID!) {
                node(id: $thread) {
                    __typename
                    ... on Thread {
                        repositoryComparison {
                            baseRepository {
                                id
                                name
                                url
                            }
                            headRepository {
                                id
                                name
                                url
                            }
                            range {
                                ...GitRevisionRangeFields
                            }
                            fileDiffs {
                                nodes {
                                    ...FileDiffFields
                                    newFile {
                                        path
                                        content
                                    }
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                                diffStat {
                                    ...DiffStatFields
                                }
                            }
                        }
                    }
                }
            }
            ${gitRevisionRangeFieldsFragment}
            ${fileDiffFieldsFragment}
            ${fileDiffHunkRangeFieldsFragment}
            ${diffStatFieldsFragment}
        `,
        { thread: thread.id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data || !data.node || data.node.__typename !== 'Thread') {
                throw new Error('thread not found')
            }
            return data.node.repositoryComparison
        })
    )
}
