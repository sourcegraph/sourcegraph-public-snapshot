import { SavedSearchFields } from '../graphql-operations'

export function urlToEditSavedSearch(savedSearch: Pick<SavedSearchFields, 'url'>): string {
    return `${savedSearch.url}/edit`
}
