import { ApolloError, ApolloQueryResult, QueryLazyOptions } from '@apollo/client'

import { useLazyQuery, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { Maybe, AffiliatedRepositoriesResult, AffiliatedRepositoriesVariables, Exact } from '../graphql-operations'

interface UseAffiliatedReposResult {
    affiliatedRepos: AffiliatedRepositoriesResult['affiliatedRepositories']['nodes'] | undefined
    loadingAffiliatedRepos: boolean
    errorAffiliatedRepos: ApolloError | undefined
    refetchAffiliatedRepos:
        | ((
              variables?:
                  | Partial<
                        Exact<{
                            user: string
                            codeHost: Maybe<string>
                            query: Maybe<string>
                        }>
                    >
                  | undefined
          ) => Promise<ApolloQueryResult<AffiliatedRepositoriesResult>>)
        | undefined
    fetchAffiliatedRepos: (
        options?:
            | QueryLazyOptions<
                  Exact<{
                      user: string
                      codeHost: Maybe<string>
                      query: Maybe<string>
                  }>
              >
            | undefined
    ) => void
}

const AFFILIATED_REPOS = gql`
    query UserAffiliatedRepositories($user: ID!, $codeHost: ID, $query: String) {
        affiliatedRepositories(user: $user, codeHost: $codeHost, query: $query) {
            nodes {
                name
                codeHost {
                    kind
                    id
                    displayName
                }
                private
            }
        }
    }
`

export const useAffiliatedRepos = (userId: string): UseAffiliatedReposResult => {
    const [trigger, { data, loading, error, refetch }] = useLazyQuery<
        AffiliatedRepositoriesResult,
        AffiliatedRepositoriesVariables
    >(AFFILIATED_REPOS, {
        variables: {
            user: userId,
            codeHost: null,
            query: null,
        },
    })

    return {
        affiliatedRepos: data?.affiliatedRepositories.nodes,
        loadingAffiliatedRepos: loading,
        errorAffiliatedRepos: error,
        refetchAffiliatedRepos: refetch,
        fetchAffiliatedRepos: trigger,
    }
}
