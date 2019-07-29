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
 * @param id The scope in which to observe the thread.
 */
export const useThreadByID = (id: GQL.ID): [typeof LOADING | GQL.IThread | null | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [threadOrError, setThreadOrError] = useState<typeof LOADING | GQL.IThread | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query ThreadByID($thread: ID!) {
                    node(id: $thread) {
                        __typename
                        ... on Thread {
                            id
                            name
                            description
                            settings
                            url
                        }
                    }
                }
            `,
            { thread: id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Thread') {
                        return null
                    }
                    return data.node
                }),
                startWith(LOADING)
            )
            .subscribe(setThreadOrError, err => setThreadOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [id, updateSequence])
    return [threadOrError, incrementUpdateSequence]
}
