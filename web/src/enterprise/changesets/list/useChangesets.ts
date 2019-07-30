import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

const changesetFields = gql`
    fragment ChangesetFields on Changeset {
        id
        number
        title
        url
    }
`

/**
 * A React hook that observes changesets queried from the GraphQL API.
 *
 * @param repository The (optional) repository in which to observe the changesets.
 */
export const useChangesets = (
    repository?: Pick<GQL.IRepository, 'id'>
): typeof LOADING | GQL.IChangesetConnection | ErrorLike => {
    const [changesets, setChangesets] = useState<typeof LOADING | GQL.IChangesetConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const results = repository
            ? queryGraphQL(
                  gql`
                      query RepositoryChangesets($repository: ID!) {
                          node(id: $repository) {
                              __typename
                              ... on Repository {
                                  changesets {
                                      nodes {
                                          ...ChangesetFields
                                      }
                                      totalCount
                                  }
                              }
                          }
                      }
                      ${changesetFields}
                  `,
                  { repository: repository.id }
              ).pipe(
                  map(dataOrThrowErrors),
                  map(data => {
                      if (!data.node || data.node.__typename !== 'Repository') {
                          throw new Error('invalid repository')
                      }
                      return data.node.changesets
                  })
              )
            : queryGraphQL(gql`
                  query Changesets {
                      changesets {
                          nodes {
                              ...ChangesetFields
                          }
                          totalCount
                      }
                  }
                  ${changesetFields}
              `).pipe(
                  map(dataOrThrowErrors),
                  map(data => data.changesets)
              )
        const subscription = results
            .pipe(startWith(LOADING))
            .subscribe(setChangesets, err => setChangesets(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository])
    return changesets
}
