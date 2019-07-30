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

export function useChangesetFileDiffs(
    changeset: Pick<GQL.IChangeset, 'id'>
): typeof LOADING | GQL.IRepositoryComparison | ErrorLike {
    const [data, setData] = useState<typeof LOADING | GQL.IRepositoryComparison | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryChangesetFileDiffs(changeset).subscribe(setData, err => setData(asError(err)))
        return () => subscription.unsubscribe()
    }, [changeset])
    return data
}

function queryChangesetFileDiffs(changeset: Pick<GQL.IChangeset, 'id'>): Observable<GQL.IRepositoryComparison> {
    return queryGraphQL(
        gql`
            query ChangesetFileDiffs($changeset: ID!) {
                node(id: $changeset) {
                    __typename
                    ... on Changeset {
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
        { changeset: changeset.id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data || !data.node || data.node.__typename !== 'Changeset') {
                throw new Error('changeset not found')
            }
            return data.node.repositoryComparison
        })
    )
}
