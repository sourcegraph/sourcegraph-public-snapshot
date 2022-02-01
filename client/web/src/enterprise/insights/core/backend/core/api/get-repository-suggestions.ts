import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../../../../backend/graphql'
import {
    RepositorySearchSuggestionsResult,
    RepositorySearchSuggestionsVariables,
} from '../../../../../../graphql-operations'
import { RepositorySuggestionData } from '../../code-insights-backend-types'

/**
 * Returns array of repository suggestions.
 *
 * @param possibleRepositoryQuery - raw string with search value for repository
 */
export const getRepositorySuggestions = (possibleRepositoryQuery: string): Promise<RepositorySuggestionData[]> =>
    requestGraphQL<RepositorySearchSuggestionsResult, RepositorySearchSuggestionsVariables>(
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
        { query: possibleRepositoryQuery }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.repositories.nodes),
            map(suggestions => suggestions.filter(suggestion => !!suggestion.name))
        )
        .toPromise()
