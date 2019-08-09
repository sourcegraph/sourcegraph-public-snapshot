import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { actorFragment, actorQuery } from '../../../actor/graphql'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a thread queried from the GraphQL API by ID.
 *
 * @param number The thread number in its repository (i.e., the `Thread.number` GraphQL
 * API field).
 */
export const useThreadByNumberInRepository = (
    repository: GQL.ID,
    number: GQL.IThread['number']
): [typeof LOADING | GQL.IThread | null | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<typeof LOADING | GQL.IThread | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query ThreadByNumberInRepository($repository: ID!, $number: String!) {
                    node(id: $repository) {
                        __typename
                        ... on Repository {
                            thread(number: $number) {
                                __typename
                                id
                                number
                                title
                                state
                                body
                                bodyHTML
                                author {
                                    ${actorQuery}
                                }
                                createdAt
                                updatedAt
                                kind
                                viewerCanUpdate
                                url
                                repository {
                                    id
                                    url
                                }
                                rules {
                                    totalCount
                                }
                                comments {
                                    totalCount
                                }
                                repositoryComparison {
                                    range {
                                        baseRevSpec {
                                            expr
                                        }
                                        headRevSpec {
                                            expr
                                        }
                                    }
                                    commits {
                                        totalCount
                                    }
                                    fileDiffs {
                                        totalCount
                                    }
                                }
                                diagnostics {
                                    totalCount
                                }
                            }
                        }
                    }
                }
                ${actorFragment}
            `,
            { repository, number }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Repository') {
                        return null
                    }
                    return data.node.thread
                })
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository, number, updateSequence])
    return [result, incrementUpdateSequence]
}
