<svelte:options immutable />

<script context="module" lang="ts">
    export type SearchResultsCapture = number
    interface ResultStateCache {
        count: number
        expanded: Set<SearchMatch>
        preview: ContentMatch | SymbolMatch | PathMatch | null
    }
    const cache = new Map<string, ResultStateCache>()

    const DEFAULT_INITIAL_ITEMS_TO_SHOW = 15
    const INCREMENTAL_ITEMS_TO_SHOW = 10
</script>

<script lang="ts">
    import { mdiCloseOctagonOutline } from '@mdi/js'
    import type { Observable } from 'rxjs'
    import { tick } from 'svelte'
    import { writable } from 'svelte/store'

    import { beforeNavigate, goto } from '$app/navigation'
    import { limitHit } from '$lib/branded'
    import Icon from '$lib/Icon.svelte'
    import { observeIntersection } from '$lib/intersection-observer'
    import type { URLQueryFilter } from '$lib/search/dynamicFilters'
    import DynamicFiltersSidebar from '$lib/search/dynamicFilters/Sidebar.svelte'
    import { createRecentSearchesStore } from '$lib/search/input/recentSearches'
    import SearchInput from '$lib/search/input/SearchInput.svelte'
    import { getQueryURL, type QueryStateStore } from '$lib/search/state'
    import PanelGroup from '$lib/wildcard/resizable-panel/PanelGroup.svelte'
    import Panel from '$lib/wildcard/resizable-panel/Panel.svelte'
    import PanelResizeHandle from '$lib/wildcard/resizable-panel/PanelResizeHandle.svelte'
    import {
        type AggregateStreamingSearchResults,
        type PathMatch,
        type SearchMatch,
        type SymbolMatch,
        type ContentMatch,
    } from '$lib/shared'

    import PreviewPanel from './PreviewPanel.svelte'
    import { getSearchResultComponent } from './searchResultFactory'
    import { setSearchResultsContext } from './searchResultsContext'
    import StreamingProgress from './StreamingProgress.svelte'

    export let stream: Observable<AggregateStreamingSearchResults>
    export let queryFromURL: string
    export let selectedFilters: URLQueryFilter[]
    export let queryState: QueryStateStore

    export function capture(): SearchResultsCapture {
        return resultContainer?.scrollTop ?? 0
    }

    export function restore(capture?: SearchResultsCapture): void {
        if (resultContainer) {
            resultContainer.scrollTop = capture ?? 0
        }
    }

    let resultContainer: HTMLElement | null = null
    const recentSearches = createRecentSearchesStore()

    $: state = $stream.state // 'loading', 'error', 'complete'
    $: results = $stream.results
    $: if (state !== 'loading') {
        recentSearches.addRecentSearch({
            query: queryFromURL,
            limitHit: limitHit($stream.progress),
            resultCount: $stream.progress.matchCount,
        })
    }

    // Logic for maintaining list state (scroll position, rendered items, open
    // items) for backwards navigation.
    $: cacheEntry = cache.get(queryFromURL)
    $: count = cacheEntry?.count ?? DEFAULT_INITIAL_ITEMS_TO_SHOW
    $: resultsToShow = results.slice(0, count)
    $: expandedSet = cacheEntry?.expanded || new Set<SearchMatch>()

    $: previewResult = writable(cacheEntry?.preview ?? null)

    setSearchResultsContext({
        isExpanded(match: SearchMatch): boolean {
            return expandedSet.has(match)
        },
        setExpanded(match: SearchMatch, expanded: boolean): void {
            if (expanded) {
                expandedSet.add(match)
            } else {
                expandedSet.delete(match)
            }
        },
        setPreview(result: ContentMatch | SymbolMatch | PathMatch | null): void {
            previewResult.set(result)
        },
        queryState,
    })
    beforeNavigate(() => {
        cache.set(queryFromURL, { count, expanded: expandedSet, preview: $previewResult })
    })

    function loadMore(event: { detail: boolean }) {
        if (event.detail) {
            count += INCREMENTAL_ITEMS_TO_SHOW
        }
    }

    // FIXME: Not a great solution since it relies on implementation details of
    // the progress component
    async function onResubmitQuery(event: SubmitEvent) {
        const target = event.currentTarget as HTMLElement | null
        const filters = Array.from(target?.querySelectorAll('[name="query"]') ?? [])
            .filter(input => (input as HTMLInputElement).checked)
            .map(input => (input as HTMLInputElement).value)
            .join(' ')
        queryState.setQuery(query => query + ' ' + filters)
        await tick()
        void goto(getQueryURL($queryState))
    }
</script>

<svelte:head>
    <title>{queryFromURL} - Sourcegraph</title>
</svelte:head>

<div class="search">
    <SearchInput {queryState} />
</div>

<div class="search-results">
    <PanelGroup id="search-results-panels">
        <Panel id="search-results-filters" order={1} defaultSize={25} maxSize={35} minSize={15}>
            <DynamicFiltersSidebar
                {selectedFilters}
                streamFilters={$stream.filters}
                searchQuery={queryFromURL}
                {state}
            />
        </Panel>
        <PanelResizeHandle />
        <Panel id="search-results-content" order={2} minSize={35}>
            <div class="results">
                <aside class="actions">
                    <StreamingProgress {state} progress={$stream.progress} on:submit={onResubmitQuery} />
                </aside>
                <div class="result-list" bind:this={resultContainer}>
                    <ol>
                        {#each resultsToShow as result, i}
                            {@const component = getSearchResultComponent(result)}
                            {#if i === resultsToShow.length - 1}
                                <li use:observeIntersection on:intersecting={loadMore}>
                                    <svelte:component this={component} {result} />
                                </li>
                            {:else}
                                <li><svelte:component this={component} {result} /></li>
                            {/if}
                        {/each}
                    </ol>
                    {#if resultsToShow.length === 0 && state !== 'loading'}
                        <div class="no-result">
                            <Icon svgPath={mdiCloseOctagonOutline} />
                            <p>No results found</p>
                        </div>
                    {/if}
                </div>
            </div>
        </Panel>

        {#if $previewResult}
            <PanelResizeHandle />
            <Panel id="search-results-file-preview" order={3} minSize={30}>
                <PreviewPanel result={$previewResult} />
            </Panel>
        {/if}
    </PanelGroup>
</div>

<style lang="scss">
    .search {
        border-bottom: 1px solid var(--border-color);
        align-self: stretch;
        padding: 0.25rem;
        // This ensures that suggestions are rendered above sticky search result headers
        z-index: 1;
    }

    .search-results {
        display: flex;
        flex: 1;
        overflow: auto;

        // Isolate everything in search results so they won't be displayed over
        // the search suggestions. Previously, hovering over separator would
        // overlap the suggestions panel.
        isolation: isolate;
    }

    .results {
        flex: 1;
        height: 100%;
        overflow: auto;
        min-height: 0;
        display: flex;
        flex-direction: column;

        .actions {
            border-bottom: 1px solid var(--border-color);
            padding: 0.5rem 0;
            padding-left: 0.25rem;
            display: flex;
            align-items: center;
            flex-shrink: 0;
        }

        .result-list {
            overflow: auto;

            ol {
                padding: 0;
                margin: 0;
                list-style: none;
            }
        }

        .no-result {
            display: flex;
            flex-direction: column;
            align-items: center;
            margin: auto;
            color: var(--text-muted);
        }
    }
</style>
