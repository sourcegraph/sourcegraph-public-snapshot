export { LazyCodeMirrorQueryInput } from './LazyCodeMirrorQueryInput'
export type { Group, Option, Completion, Target, Command, Source } from './suggestions'
export { getEditorConfig } from './suggestions'
import FilterOption from './FilterSuggestion.svelte'
import SearchQueryOption from './SearchQueryOption.svelte'
import { CustomRenderer } from './suggestions'

// Some type casting dance to make TypeScript happy...
const SearchQueryOptionTyped = SearchQueryOption as CustomRenderer
export { SearchQueryOptionTyped as SearchQueryOption }
const FilterOptionTyped = FilterOption as CustomRenderer
export { FilterOptionTyped as FilterOption }
