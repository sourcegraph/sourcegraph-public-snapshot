import { gql, useQuery } from '@apollo/client'

import { PackageSearchSuggestionsResult } from '../../../graphql-operations'

const GET_PACKAGE_SUGGESTIONS = gql`
    query PackageSearchSuggestions($query: String) {
        packageRepoReferences(first: 10, name: $query) {
            nodes {
                id
                name
            }
        }
    }
`

interface UsePackageSuggestionsProps {
    search: string
    disable?: boolean
}

interface UsePackageSuggestionsResult {
    suggestions: string[]
    loading: boolean
}

export function usePackageSuggestions(props: UsePackageSuggestionsProps): UsePackageSuggestionsResult {
    const { search, disable = false } = props

    const {
        data: currentData,
        previousData,
        loading,
    } = useQuery<PackageSearchSuggestionsResult>(GET_PACKAGE_SUGGESTIONS, {
        skip: disable,
        fetchPolicy: 'cache-and-network',
        variables: { query: search === '' ? null : search },
    })

    const data = currentData ?? previousData

    return {
        suggestions: data?.packageRepoReferences.nodes.map(suggestion => suggestion.name) ?? [],
        loading,
    }
}
