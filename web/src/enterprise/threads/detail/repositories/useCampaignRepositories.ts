import { useEffect, useState } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'
import { fileDiffFieldsFragment } from '../../../../repo/compare/RepositoryCompareDiffPage'
import { gitRevisionRangeFieldsFragment } from '../../../../repo/compare/RepositoryCompareOverviewPage'

const LOADING: 'loading' = 'loading'

export function useThreadRepositories(
    thread: Pick<GQL.IThread, 'id'>
): typeof LOADING | GQL.IRepositoryComparison[] | ErrorLike {
    const [data, setData] = useState<typeof LOADING | GQL.IRepositoryComparison[] | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryThreadRepositories(thread).subscribe(setData, err => setData(asError(err)))
        return () => subscription.unsubscribe()
    }, [thread])
    return data
}

function queryThreadRepositories(thread: Pick<GQL.IThread, 'id'>): Observable<GQL.IRepositoryComparison[]> {
    return queryGraphQL(
        gql`
            query ThreadRepositories($thread: ID!) {
                node(id: $thread) {
                    __typename
                    ... on Thread {
                        repositoryComparisons {
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
                            commits {
                                nodes {
                                    ...GitCommitFields
                                }
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
            }
            ${gitRevisionRangeFieldsFragment}
            ${gitCommitFragment}
            ${fileDiffFieldsFragment}
        `,
        { thread }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data || !data.node || data.node.__typename !== 'Thread') {
                throw new Error('thread not found')
            }
            return data.node.repositoryComparisons
        })
    )
}
