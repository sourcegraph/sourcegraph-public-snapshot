import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import { RepositorySearchSuggestionsResult, RepositorySearchSuggestionsVariables } from '../../../../graphql-operations'

export interface RepositorySuggestion {
    /**
     * Repository id.
     */
    id: string
    /**
     * Repository name.
     */
    name: string
}

/**
 * Fetch repository suggestions for repositories/repository input.
 * Used in code insight creation UI form repository field.
 *
 * @param possibleRepository - string with possible repository name
 */
export function fetchRepositorySuggestions(possibleRepository: string): Observable<RepositorySuggestion[]> {
    return requestGraphQL<RepositorySearchSuggestionsResult, RepositorySearchSuggestionsVariables>(
        gql`
            query RepositorySearchSuggestions($query: String!) {
                repositories(first: 5, query: $query) {
                    nodes {
                        id
                        name
                    }
                }
            }
        `,
        { query: possibleRepository }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repositories.nodes),
        map(suggestions => suggestions.filter(suggestion => !!suggestion.name))
    )
}
