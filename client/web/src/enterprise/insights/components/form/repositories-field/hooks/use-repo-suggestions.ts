import { gql, useQuery } from '@apollo/client'

import { asError } from '@sourcegraph/common'

import { RepositorySearchSuggestionsResult } from '../../../../../../graphql-operations'

const GET_REPOSITORY_SUGGESTION = gql`
    query RepositorySearchSuggestions($query: String!) {
        repositories(first: 5, query: $query) {
            nodes {
                id
                name
            }
        }
    }
`

interface UseRepoSuggestionsProps {
    search: string | null
    disable?: boolean
}

export interface RepositorySuggestionData {
    id: string
    name: string
}

export function useRepoSuggestions(props: UseRepoSuggestionsProps): RepositorySuggestionData[] | Error | undefined {
    const { search, disable = false } = props

    const { data, loading, error } = useQuery<RepositorySearchSuggestionsResult>(GET_REPOSITORY_SUGGESTION, {
        skip: disable || !search,
        fetchPolicy: 'cache-first',
        variables: { query: search },
    })

    if (error) {
        return asError(error)
    }

    if (loading) {
        return undefined
    }

    if (data) {
        return data.repositories.nodes.filter(suggestion => !!suggestion.name)
    }

    return []
}
