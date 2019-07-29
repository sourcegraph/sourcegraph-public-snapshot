import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a thread queried from the GraphQL API by ID.
 *
 * @param idInRepository The thread ID in its repository (i.e., the `Thread.idInRepository` GraphQL
 * API field).
 */
export const useThreadByIDInRepository = (
    repository: GQL.ID,
    threadIDInRepository: GQL.IThread['idInRepository']
): [typeof LOADING | GQL.IThread | null | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<typeof LOADING | GQL.IThread | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query ThreadByIDInRepository($repository: ID!, $threadIDInRepository: String!) {
                    node(id: $repository) {
                        __typename
                        ... on Repository {
                            thread(idInRepository: $threadIDInRepository) {
                                id
                                idInRepository
                                title
                                url
                            }
                        }
                    }
                }
            `,
            { repository, threadIDInRepository }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Repository') {
                        return null
                    }
                    return data.node.thread
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository, threadIDInRepository, updateSequence])
    return [result, incrementUpdateSequence]
}
