import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes all threads in a thread (queried from the GraphQL API).
 *
 * @param thread The thread whose threads to observe.
 */
export const useThreadThreads = (
    thread: Pick<GQL.IThread, 'id'>
): [typeof LOADING | GQL.IThreadConnection | ErrorLike, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IThreadConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query ThreadThreads($thread: ID!) {
                    node(id: $thread) {
                        __typename
                        ... on Thread {
                            threads {
                                nodes {
                                    id
                                    title
                                    url
                                    status
                                    type
                                }
                                totalCount
                            }
                        }
                    }
                }
            `,
            { thread: thread.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Thread') {
                        throw new Error('not a thread')
                    }
                    return data.node.threads
                }),
                startWith(LOADING)
            )
            .subscribe(setThreadsOrError, err => setThreadsOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [thread, updateSequence])
    return [threadsOrError, incrementUpdateSequence]
}
