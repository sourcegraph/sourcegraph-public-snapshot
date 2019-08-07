import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'
import { ThreadFragment } from '../util/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes threads queried from the GraphQL API.
 *
 * @param repository The (optional) repository in which to observe the threads.
 */
export const useThreads = (
    repository?: Pick<GQL.IRepository, 'id'>
): typeof LOADING | GQL.IThreadConnection | ErrorLike => {
    const [threads, setThreads] = useState<typeof LOADING | GQL.IThreadConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const results = repository
            ? queryGraphQL(
                  gql`
                      query RepositoryThreads($repository: ID!) {
                          node(id: $repository) {
                              __typename
                              ... on Repository {
                                  threads {
                                      nodes {
                                          ...ThreadFragment
                                      }
                                      totalCount
                                  }
                              }
                          }
                      }
                      ${ThreadFragment}
                  `,
                  { repository: repository.id }
              ).pipe(
                  map(dataOrThrowErrors),
                  map(data => {
                      if (!data.node || data.node.__typename !== 'Repository') {
                          throw new Error('invalid repository')
                      }
                      return data.node.threads
                  })
              )
            : queryGraphQL(gql`
                  query Threads {
                      threads {
                          nodes {
                              ...ThreadFragment
                          }
                          totalCount
                      }
                  }
                  ${ThreadFragment}
              `).pipe(
                  map(dataOrThrowErrors),
                  map(data => data.threads)
              )
        const subscription = results.pipe(startWith(LOADING)).subscribe(setThreads, err => setThreads(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository])
    return threads
}
