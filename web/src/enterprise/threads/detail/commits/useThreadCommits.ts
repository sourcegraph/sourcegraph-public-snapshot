import { useEffect, useState } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'

const LOADING: 'loading' = 'loading'

export function useThreadCommits(
    thread: Pick<GQL.IThread, 'id'>
): typeof LOADING | GQL.IGitCommit[] | ErrorLike {
    const [data, setData] = useState<typeof LOADING | GQL.IGitCommit[] | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryThreadCommits(thread).subscribe(setData, err => setData(asError(err)))
        return () => subscription.unsubscribe()
    }, [thread])
    return data
}

function queryThreadCommits(thread: Pick<GQL.IThread, 'id'>): Observable<GQL.IGitCommit[]> {
    return queryGraphQL(
        gql`
            query ThreadCommits($thread: ID!) {
                node(id: $thread) {
                    __typename
                    ... on Thread {
                        repositoryComparison {
                            commits {
                                nodes {
                                    ...GitCommitFields
                                }
                            }
                        }
                    }
                }
            }
            ${gitCommitFragment}
        `,
        { thread: thread.id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data || !data.node || data.node.__typename !== 'Thread') {
                throw new Error('thread not found')
            }
            if (!data.node.repositoryComparison) {
                throw new Error('thread has no repository comparison')
            }
            return data.node.repositoryComparison.commits.nodes
        })
    )
}
