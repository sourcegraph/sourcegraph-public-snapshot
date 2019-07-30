import { useEffect, useState } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'

const LOADING: 'loading' = 'loading'

export function useChangesetCommits(
    changeset: Pick<GQL.IChangeset, 'id'>
): typeof LOADING | GQL.IGitCommit[] | ErrorLike {
    const [data, setData] = useState<typeof LOADING | GQL.IGitCommit[] | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryChangesetCommits(changeset).subscribe(setData, err => setData(asError(err)))
        return () => subscription.unsubscribe()
    }, [changeset])
    return data
}

function queryChangesetCommits(changeset: Pick<GQL.IChangeset, 'id'>): Observable<GQL.IGitCommit[]> {
    return queryGraphQL(
        gql`
            query ChangesetCommits($changeset: ID!) {
                node(id: $changeset) {
                    __typename
                    ... on Changeset {
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
        { changeset: changeset.id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data || !data.node || data.node.__typename !== 'Changeset') {
                throw new Error('changeset not found')
            }
            if (!data.node.repositoryComparison) {
                throw new Error('changeset has no repository comparison')
            }
            return data.node.repositoryComparison.commits.nodes
        })
    )
}
