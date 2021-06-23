import { fetchRepositoriesBySearch } from '../requests/fetch-repositories-by-search'

/**
 * Get list of resolved repositories from the search API.
 *
 * @param query - search query
 */
export const getResolvedSearchRepositories = (query: string): Promise<string[]> =>
    fetchRepositoriesBySearch(query).toPromise()
