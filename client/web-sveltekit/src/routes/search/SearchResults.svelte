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
    import type { Observable } from 'rxjs'
    import { onMount, tick } from 'svelte'
    import { writable } from 'svelte/store'

    import { beforeNavigate, goto } from '$app/navigation'
    import { limitHit } from '$lib/branded'
    import Icon2 from '$lib/Icon2.svelte'
    import { observeIntersection } from '$lib/intersection-observer'
    import GlobalHeaderPortal from '$lib/navigation/GlobalHeaderPortal.svelte'
    import type { URLQueryFilter } from '$lib/search/dynamicFilters'
    import DynamicFiltersSidebar from '$lib/search/dynamicFilters/Sidebar.svelte'
    import { createRecentSearchesStore } from '$lib/search/input/recentSearches'
    import SearchInput from '$lib/search/input/SearchInput.svelte'
    import { getQueryURL, type QueryStateStore } from '$lib/search/state'
    import {
        TELEMETRY_SEARCH_SOURCE_TYPE,
        type AggregateStreamingSearchResults,
        type PathMatch,
        type SearchMatch,
        type SymbolMatch,
        type ContentMatch,
    } from '$lib/shared'
    import { TELEMETRY_RECORDER } from '$lib/telemetry'
    import Panel from '$lib/wildcard/resizable-panel/Panel.svelte'
    import PanelGroup from '$lib/wildcard/resizable-panel/PanelGroup.svelte'
    import PanelResizeHandle from '$lib/wildcard/resizable-panel/PanelResizeHandle.svelte'

    import PreviewPanel from './PreviewPanel.svelte'
    import SearchAlert from './SearchAlert.svelte'
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

    onMount(() => {
        TELEMETRY_RECORDER.recordEvent('search.results', 'view')
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

    function handleResultCopy(): void {
        TELEMETRY_RECORDER.recordEvent('search.result.area', 'copy')
    }

    function handleSearchResultClick(index: number): void {
        TELEMETRY_RECORDER.recordEvent('search.result.area', 'click', {
            metadata: {
                index,
                resultsLength: results.length,
            },
        })
    }

    function handleSubmit() {
        TELEMETRY_RECORDER.recordEvent('search', 'submit', {
            metadata: { source: TELEMETRY_SEARCH_SOURCE_TYPE['nav'] },
        })
    }
</script>

<GlobalHeaderPortal>
    <div class="search-header">
        <SearchInput {queryState} size="compat" onSubmit={handleSubmit} />
    </div>
</GlobalHeaderPortal>

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
                    {#if $stream.alert}
                        <div class="message-container">
                            <SearchAlert alert={$stream.alert} />
                        </div>
                    {/if}
                    <!--
                        TODO: Address accessibility issues
                        1. A11y: visible, non-interactive elements with an on:click event
                           must be accompanied by an on:keydown, on:keyup, or on:keypress event.
                        2. A11y: Non-interactive element <ol> should not be assigned mouse
                           or keyboard event listeners.
                    -->
                    <ol on:copy={handleResultCopy}>
                        {#each resultsToShow as result, i}
                            {@const component = getSearchResultComponent(result)}
                            {#if i === resultsToShow.length - 1}
                                <li
                                    use:observeIntersection
                                    on:intersecting={loadMore}
                                    on:click={() => handleSearchResultClick(i)}
                                >
                                    <svelte:component this={component} {result} />
                                </li>
                            {:else}
                                <li on:click={() => handleSearchResultClick(i)}>
                                    <svelte:component this={component} {result} />
                                </li>
                            {/if}
                        {/each}
                    </ol>
                    {#if resultsToShow.length === 0 && state !== 'loading'}
                        <div class="message-container">
                            <Icon2 icon={ILucideOctagonX} aria-hidden="true" />
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
    .search-header {
        width: 100%;
        // This ensures that the search suggestions panel is displayed above the
        // search results panel.
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
            padding: 0.5rem;
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

        .message-container {
            display: flex;
            flex-direction: column;
            align-items: center;
            margin: auto;
            color: var(--text-muted);
            margin: 2rem;
        }
    }
</style>
