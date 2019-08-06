import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { actorFragment, actorQuery } from '../../../actor/graphql'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a issue queried from the GraphQL API by ID.
 *
 * @param number The issue number in its repository (i.e., the `Issue.number` GraphQL
 * API field).
 */
export const useIssueByNumberInRepository = (
    repository: GQL.ID,
    number: GQL.IIssue['number']
): [typeof LOADING | GQL.IIssue | null | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<typeof LOADING | GQL.IIssue | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query IssueByNumberInRepository($repository: ID!, $number: String!) {
                    node(id: $repository) {
                        __typename
                        ... on Repository {
                            issue(number: $number) {
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
                                viewerCanUpdate
                                url
                                repository {
                                    url
                                }
                                diagnosticsData
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
                    return data.node.issue
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository, number, updateSequence])
    return [result, incrementUpdateSequence]
}
