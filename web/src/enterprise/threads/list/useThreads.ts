import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes threads queried from the GraphQL API.
 *
 * @param namespace The (optional) namespace in which to observe the threads defined.
 */
export const useThreads = (
    namespace?: Pick<GQL.INamespace, 'id'>
): typeof LOADING | GQL.IThreadConnection | ErrorLike => {
    const [threads, setThreads] = useState<typeof LOADING | GQL.IThreadConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const results = namespace
            ? queryGraphQL(
                  gql`
                      query ThreadsDefinedInNamespace($namespace: ID!) {
                          namespace(id: $namespace) {
                              threads {
                                  nodes {
                                      id
                                      name
                                      description
                                      url
                                  }
                                  totalCount
                              }
                          }
                      }
                  `,
                  { namespace: namespace.id }
              ).pipe(
                  map(dataOrThrowErrors),
                  map(data => {
                      if (!data.namespace) {
                          throw new Error('not a namespace')
                      }
                      return data.namespace.threads
                  })
              )
            : queryGraphQL(gql`
                  query Threads {
                      threads {
                          nodes {
                              id
                              name
                              description
                              url
                          }
                          totalCount
                      }
                  }
              `).pipe(
                  map(dataOrThrowErrors),
                  map(data => data.threads)
              )
        const subscription = results.pipe(startWith(LOADING)).subscribe(setThreads, err => setThreads(asError(err)))
        return () => subscription.unsubscribe()
    }, [namespace])
    return threads
}
