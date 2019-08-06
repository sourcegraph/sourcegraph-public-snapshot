import { useEffect, useState } from 'react'
import { map, startWith } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'

const LOADING: 'loading' = 'loading'

const issueFields = gql`
    fragment IssueFields on Issue {
        __typename
        id
        number
        title
        state
        url
    }
`

/**
 * A React hook that observes issues queried from the GraphQL API.
 *
 * @param repository The (optional) repository in which to observe the issues.
 */
export const useIssues = (
    repository?: Pick<GQL.IRepository, 'id'>
): typeof LOADING | GQL.IIssueConnection | ErrorLike => {
    const [issues, setIssues] = useState<typeof LOADING | GQL.IIssueConnection | ErrorLike>(LOADING)
    useEffect(() => {
        const results = repository
            ? queryGraphQL(
                  gql`
                      query RepositoryIssues($repository: ID!) {
                          node(id: $repository) {
                              __typename
                              ... on Repository {
                                  issues {
                                      nodes {
                                          ...IssueFields
                                      }
                                      totalCount
                                  }
                              }
                          }
                      }
                      ${issueFields}
                  `,
                  { repository: repository.id }
              ).pipe(
                  map(dataOrThrowErrors),
                  map(data => {
                      if (!data.node || data.node.__typename !== 'Repository') {
                          throw new Error('invalid repository')
                      }
                      return data.node.issues
                  })
              )
            : queryGraphQL(gql`
                  query Issues {
                      issues {
                          nodes {
                              ...IssueFields
                          }
                          totalCount
                      }
                  }
                  ${issueFields}
              `).pipe(
                  map(dataOrThrowErrors),
                  map(data => data.issues)
              )
        const subscription = results
            .pipe(startWith(LOADING))
            .subscribe(setIssues, err => setIssues(asError(err)))
        return () => subscription.unsubscribe()
    }, [repository])
    return issues
}
