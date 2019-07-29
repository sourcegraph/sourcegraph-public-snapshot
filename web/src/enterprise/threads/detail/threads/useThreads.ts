import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes all threads (queried from the GraphQL API).
 */
export const useThreads = (): typeof LOADING | GQL.IThreadConnection | ErrorLike => {
    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IThreadConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query Threads {
                    threads {
                        nodes {
                            id
                            idWithoutKind
                            title
                            url
                            status
                            type
                        }
                        totalCount
                    }
                }
            `
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => data.threads),
                startWith(LOADING)
            )
            .subscribe(setThreadsOrError, err => setThreadsOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [])
    return threadsOrError
}
