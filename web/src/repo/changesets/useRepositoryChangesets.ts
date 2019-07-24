import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a repository's changesets (queried from the GraphQL API).
 *
 * @param repository The repository whose changesets to observe.
 */
export const useRepositoryChangesets = (
    repository: Pick<GQL.IRepository, 'id'>
): typeof LOADING | GQL.IChangesetConnection | ErrorLike => {
    const [changesetsOrError, setChangesetsOrError] = useState<typeof LOADING | GQL.IChangesetConnection | ErrorLike>(
        LOADING
    )
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query RepositoryChangesets($repository: ID!) {
                    node(id: $repository) {
                        __typename
                        ... on Repository {
                            changesets {
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
                    return data.node.changesets
                }),
                startWith(LOADING)
            )
            .subscribe(setChangesetsOrError, err => setChangesetsOrError(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository])
    return changesetsOrError
}
