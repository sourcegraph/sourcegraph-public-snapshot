import { IRepoGroup, ISearchContext, SearchSuggestion as DynamicSearchSuggestion } from '../../graphql/schema'

export type SearchSuggestion = DynamicSearchSuggestion | IRepoGroup | ISearchContext
