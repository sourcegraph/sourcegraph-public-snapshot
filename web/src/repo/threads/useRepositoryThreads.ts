import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a repository's threads (queried from the GraphQL API).
 *
 * @param repository The repository whose threads to observe.
 */
export const useRepositoryThreads = (
    repository: Pick<GQL.IRepository, 'id'>
): typeof LOADING | GQL.IThreadConnection | ErrorLike => {
    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IThreadConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query RepositoryThreads($repository: ID!) {
                    node(id: $repository) {
                        __typename
                        ... on Repository {
                            threads {
                                nodes {
                                    title
                                    externalURL
                                }
                                totalCount
                            }
                        }
                    }
                }
            `,
            { repository: repository.id }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => {
                    if (!data.node || data.node.__typename !== 'Repository') {
                        throw new Error('not a repository')
                    }
                    return data.node.threads
                }),
                startWith(LOADING)
            )
            .subscribe(setThreadsOrError, err => setThreadsOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository])
    return threadsOrError
}
