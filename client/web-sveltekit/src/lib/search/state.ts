import { writable, type Readable } from 'svelte/store'

import { goto } from '$app/navigation'
import { SearchPatternType } from '$lib/graphql-operations'
import { buildSearchURLQuery, type SettingsCascade } from '$lib/shared'
import { defaultSearchModeFromSettings } from '$lib/web'

// Defined in @sourcegraph/shared/src/search/searchQueryState.tsx
export enum SearchMode {
    Precise = 0,
    SmartSearch = 1 << 0,
}

type Update<T> = T | ((value: T) => T)

interface Options {
    readonly caseSensitive: boolean
    readonly regularExpression: boolean
    readonly patternType: SearchPatternType
    readonly searchMode: SearchMode
    readonly query: string
    readonly searchContext: string
}

type QuerySettings = Pick<
    SettingsCascade['final'],
    'search.defaultCaseSensitive' | 'search.defaultPatternType' | 'search.defaultMode'
> | null
export type QueryOptions = Pick<Options, 'patternType' | 'caseSensitive' | 'searchMode' | 'searchContext'>

export class QueryState {
    private defaultCaseSensitive = false
    private defaultPatternType = SearchPatternType.standard
    private defaultSearchMode = SearchMode.Precise
    private defaultQuery = ''
    private defaultSearchContext = 'global'

    private constructor(public readonly options: Partial<Options>, public settings: QuerySettings) {}

    public static init(options: Partial<Options>, settings: QuerySettings): QueryState {
        return new QueryState(options, settings)
    }

    public get caseSensitive(): boolean {
        return this.options.caseSensitive ?? this.settings?.['search.defaultCaseSensitive'] ?? this.defaultCaseSensitive
    }

    public get patternType(): SearchPatternType {
        return (
            this.options.patternType ??
            (this.settings?.['search.defaultPatternType'] as SearchPatternType) ??
            this.defaultPatternType
        )
    }

    public get searchMode(): SearchMode {
        return (
            // {final: this.settings, subjects} is a workaround to make our
            // settings representation work with defaultSearchModeFromSettings
            this.options.searchMode ??
            (this.settings ? defaultSearchModeFromSettings({ final: this.settings, subjects: [] }) : null) ??
            this.defaultSearchMode
        )
    }

    public get query(): string {
        return this.options.query ?? this.defaultQuery
    }

    public get searchContext(): string {
        return this.options.searchContext ?? this.defaultSearchContext
    }

    public setQuery(newQuery: Update<string>): QueryState {
        const query = typeof newQuery === 'function' ? newQuery(this.query) : newQuery
        return new QueryState({ ...this.options, query }, this.settings)
    }

    public setCaseSensitive(caseSensitive: Update<boolean>): QueryState {
        return new QueryState(
            {
                ...this.options,
                caseSensitive: typeof caseSensitive === 'function' ? caseSensitive(this.caseSensitive) : caseSensitive,
            },
            this.settings
        )
    }

    public setPatternType(patternType: Update<SearchPatternType>): QueryState {
        return new QueryState(
            {
                ...this.options,
                patternType: typeof patternType === 'function' ? patternType(this.patternType) : patternType,
            },
            this.settings
        )
    }

    public setMode(mode: SearchMode): QueryState {
        return new QueryState({ ...this.options, searchMode: mode }, this.settings)
    }

    public setSettings(settings: QuerySettings): QueryState {
        return new QueryState(this.options, settings)
    }
}

export interface QueryStateStore extends Readable<QueryState> {
    setQuery(update: Update<string>): void
    setCaseSensitive(update: Update<boolean>): void
    setPatternType(update: Update<SearchPatternType>): void
    setSettings(settings: QuerySettings): void
    setMode(mode: SearchMode): void
    set(options: Partial<Options>): void
}

export function queryStateStore(initial: Partial<Options> = {}, settings: QuerySettings): QueryStateStore {
    const { subscribe, update } = writable<QueryState>(QueryState.init(initial, settings))
    return {
        subscribe,
        setQuery(newQuery) {
            update(state => state.setQuery(newQuery))
        },
        setCaseSensitive(caseSensitive) {
            update(state => state.setCaseSensitive(caseSensitive))
        },
        setPatternType(patternType) {
            update(state => state.setPatternType(patternType))
        },
        setSettings(settings) {
            update(state => state.setSettings(settings))
        },
        setMode(mode) {
            update(state => state.setMode(mode))
        },
        set(options: Partial<Options>) {
            update(state => QueryState.init({ ...state.options, ...options }, state.settings))
        },
    }
}

export function getQueryURL(
    queryState: Pick<QueryState, 'searchMode' | 'query' | 'caseSensitive' | 'patternType' | 'searchContext'>
): string {
    const searchQueryParameter = buildSearchURLQuery(
        queryState.query,
        queryState.patternType,
        queryState.caseSensitive,
        queryState.searchContext,
        queryState.searchMode
    )

    return '/search?' + searchQueryParameter
}

export function submitSearch(
    queryState: Pick<QueryState, 'searchMode' | 'query' | 'caseSensitive' | 'patternType' | 'searchContext'>
): void {
    void goto(getQueryURL(queryState))
}
