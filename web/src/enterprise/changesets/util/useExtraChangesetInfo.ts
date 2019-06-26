import { useState } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'
import { gitCommitFragment } from '../../../repo/commits/RepositoryCommitsPage'
import {
    diffStatFieldsFragment,
    fileDiffFieldsFragment,
    fileDiffHunkRangeFieldsFragment,
} from '../../../repo/compare/RepositoryCompareDiffPage'
import { gitRevisionRangeFieldsFragment } from '../../../repo/compare/RepositoryCompareOverviewPage'
import { useEffectAsync } from '../../../util/useEffectAsync'

const LOADING: 'loading' = 'loading'

export function useExtraChangesetInfo(
    thread: Pick<GQL.IDiscussionThread, 'id'> | typeof LOADING | ErrorLike
): typeof LOADING | GQL.IChangeset | ErrorLike {
    const [changesetOrError, setChangesetOrError] = useState<typeof LOADING | GQL.IChangeset | ErrorLike>(LOADING)
    useEffectAsync(async () => {
        if (thread === LOADING || isErrorLike(thread)) {
            setChangesetOrError(LOADING)
        } else {
            try {
                // TODO!(sqs)
                setChangesetOrError(await queryChangeset(thread.id).toPromise())
            } catch (err) {
                setChangesetOrError(asError(err))
            }
        }
    }, [thread])
    return changesetOrError
}

function queryChangeset(threadID: string): Observable<GQL.IChangeset> {
    return queryGraphQL(
        gql`
            query Changeset($threadID: ID!, $first: Int!) {
                node(id: $threadID) {
                    __typename
                    ... on DiscussionThread {
                        changeset {
                            repositories {
                                id
                                name
                                url
                            }
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
                                commits(first: $first) {
                                    nodes {
                                        ...GitCommitFields
                                    }
                                    pageInfo {
                                        hasNextPage
                                    }
                                }
                                fileDiffs(first: $first) {
                                    nodes {
                                        ...FileDiffFields
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
            }
            ${gitRevisionRangeFieldsFragment}
            ${gitCommitFragment}
            ${fileDiffFieldsFragment}
            ${fileDiffHunkRangeFieldsFragment}
            ${diffStatFieldsFragment}
        `,
        { threadID, first: 999999 /* TODO!(sqs) */ }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node || data.node.__typename !== 'DiscussionThread' || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            if (!data.node.changeset) {
                throw new Error('thread is not a changeset')
            }
            return data.node.changeset
        })
    )
}
