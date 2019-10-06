import { useCallback, useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'

export const ThreadDiagnosticConnectionFragment = gql`
    fragment ThreadDiagnosticConnectionFragment on ThreadDiagnosticConnection {
        edges {
            id
            diagnostic {
                type
                data
            }
            viewerCanUpdate
        }
        totalCount
    }
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.IThreadDiagnosticConnection | ErrorLike

/**
 * A React hook that observes all diagnostics for a thread (queried from the GraphQL API).
 *
 * @param thread The thread whose diagnostics to observe.
 */
export const useThreadDiagnostics = (thread: Pick<GQL.IThread, 'id'>): [Result, () => void] => {
    const [updateSequence, setUpdateSequence] = useState(0)
    const incrementUpdateSequence = useCallback(() => setUpdateSequence(updateSequence + 1), [updateSequence])

    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query ThreadDiagnostics($thread: ID!) {
                    node(id: $thread) {
                        __typename
                        ... on Thread {
                            diagnostics {
                                ...ThreadDiagnosticConnectionFragment
                            }
                        }
                    }
                }
                ${ThreadDiagnosticConnectionFragment}
            `,
            { thread: thread.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Thread') {
                        throw new Error('not a thread')
                    }
                    return data.node.diagnostics
                }),
                startWith(LOADING)
            )
            .subscribe(setResult, err => setResult(asError(err)))
        return () => subscription.unsubscribe()
    }, [thread, updateSequence])
    return [result, incrementUpdateSequence]
}
