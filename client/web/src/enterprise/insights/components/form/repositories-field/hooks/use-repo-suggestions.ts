import { gql, useQuery } from '@apollo/client'

import type { RepositorySearchSuggestionsResult } from '../../../../../../graphql-operations'

const GET_REPOSITORY_SUGGESTION = gql`
    query RepositorySearchSuggestions($query: String) {
        repositories(first: 10, query: $query) {
            nodes {
                id
                name
            }
        }
    }
`

interface UseRepoSuggestionsProps {
    search: string
    disable?: boolean
    selectedRepositories: string[]
}

interface UseRepoSuggestionsResult {
    suggestions: string[]
    loading: boolean
}

export function useRepoSuggestions(props: UseRepoSuggestionsProps): UseRepoSuggestionsResult {
    const { search, disable = false, selectedRepositories } = props

    const {
        data: currentData,
        previousData,
        loading,
    } = useQuery<RepositorySearchSuggestionsResult>(GET_REPOSITORY_SUGGESTION, {
        skip: disable,
        fetchPolicy: 'cache-and-network',
        variables: { query: search === '' ? null : search },
    })

    const data = currentData ?? previousData

    return {
        suggestions:
            data?.repositories.nodes
                ?.filter(suggestion => !!suggestion.name && !selectedRepositories.includes(suggestion.name))
                .map(suggestion => suggestion.name) ?? [],
        loading,
    }
}
