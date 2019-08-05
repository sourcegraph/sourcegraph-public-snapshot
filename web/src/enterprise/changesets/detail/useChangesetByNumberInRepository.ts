import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a changeset queried from the GraphQL API by ID.
 *
 * @param number The changeset number in its repository (i.e., the `Changeset.number` GraphQL
 * API field).
 */
export const useChangesetByNumberInRepository = (
    repository: GQL.ID,
    number: GQL.IChangeset['number']
): [typeof LOADING | GQL.IChangeset | null | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<typeof LOADING | GQL.IChangeset | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query ChangesetByNumberInRepository($repository: ID!, $number: String!) {
                    node(id: $repository) {
                        __typename
                        ... on Repository {
                            changeset(number: $number) {
                                id
                                number
                                title
                                body
                                bodyHTML
                                author {
                                    ... on User {
                                        displayName
                                        username
                                        url
                                    }
                                }
                                createdAt
                                updatedAt
                                viewerCanUpdate
                                url
                                repository {
                                    url
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
                                }
                            }
                        }
                    }
                }
            `,
            { repository, number }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Repository') {
                        return null
                    }
                    return data.node.changeset
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository, number, updateSequence])
    return [result, incrementUpdateSequence]
}
