<script context="module" lang="ts">
    interface Cache {
        count: number
        expanded: Set<SearchMatch>
    }
    const cache = new Map<string, Cache>()

    export interface SearchResultsContext {
        isExpanded(match: SearchMatch): boolean
        setExpanded(match: SearchMatch, expanded: boolean): void
        queryState: QueryStateStore
    }

    const CONTEXT_KEY = 'search-result'

    export function getSearchResultsContext(): SearchResultsContext {
        return getContext(CONTEXT_KEY)
    }

    export function setSearchResultsContext(context: SearchResultsContext): SearchResultsContext {
        return setContext(CONTEXT_KEY, context)
    }

    const DEFAULT_INITIAL_ITEMS_TO_SHOW = 15
    const INCREMENTAL_ITEMS_TO_SHOW = 10
</script>

<script lang="ts">
    import type { Observable } from 'rxjs'
    import { getContext, setContext, tick } from 'svelte'

    import { beforeNavigate } from '$app/navigation'
    import { preserveScrollPosition } from '$lib/app'
    import { observeIntersection } from '$lib/intersection-observer'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import SearchBox from '$lib/search/SearchBox.svelte'
    import { searchTypes } from '$lib/search/sidebar'
    import { submitSearch, type QueryStateStore } from '$lib/search/state'
    import type { SidebarFilter } from '$lib/search/utils'
    import { SearchSidebarSectionID, type AggregateStreamingSearchResults, type SearchMatch } from '$lib/shared'

    import Section from './SidebarSection.svelte'
    import StreamingProgress from './StreamingProgress.svelte'
    import { getSearchResultComponent } from './searchResultFactory'

    export let stream: Observable<AggregateStreamingSearchResults | undefined>
    export let queryFromURL: string
    export let queryState: QueryStateStore

    let resultContainer: HTMLElement | null = null
    let searchInput: SearchBox

    $: progress = $stream?.progress
    // NOTE: done is present but apparently not officially exposed. However
    // $stream.state is always "loading". Need to look into this.
    $: loading = !progress?.done
    $: results = $stream?.results
    $: filters = $stream?.filters
    $: langFilters =
        filters
            ?.filter(filter => filter.kind === 'lang')
            .map((filter): SidebarFilter => ({ ...filter, runImmediately: true })) ?? []

    // Logic for maintaining list state (scroll position, rendered items, open
    // items) for backwards navigation.
    $: cacheEntry = cache.get(queryFromURL)
    $: count = cacheEntry?.count ?? DEFAULT_INITIAL_ITEMS_TO_SHOW
    $: resultsToShow = results ? results.slice(0, count) : null
    $: expandedSet = cacheEntry?.expanded || new Set<SearchMatch>()

    let scrollTop: number = 0
    preserveScrollPosition(
        position => (scrollTop = position ?? 0),
        () => resultContainer?.scrollTop
    )
    $: if (resultContainer) {
        resultContainer.scrollTop = scrollTop ?? 0
    }
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
        queryState,
    })
    beforeNavigate(() => {
        cache.set(queryFromURL, { count, expanded: expandedSet })
    })

    function loadMore(event: { detail: boolean }) {
        if (event.detail) {
            count += INCREMENTAL_ITEMS_TO_SHOW
        }
    }

    async function updateQuery(event: MouseEvent) {
        const element = event.currentTarget as HTMLElement
        // TODO: Replace / update query; editor hints; etc
        queryState.setQuery(query => query + ' ' + element.dataset.value)
        if (element.dataset.run) {
            await tick()
            submitSearch($queryState)
        } else {
            searchInput.focus()
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
        submitSearch($queryState)
    }
</script>

<section>
    <div class="search">
        <SearchBox bind:this={searchInput} {queryState} />
    </div>

    <div class="results" bind:this={resultContainer}>
        <div class="scroll-container">
            {#if !$stream || loading}
                <div class="spinner">
                    <LoadingSpinner />
                </div>
            {:else if !loading && resultsToShow}
                <div class="main">
                    <aside class="stats mb-2">
                        {#if progress}
                            <StreamingProgress {progress} on:submit={onResubmitQuery} />
                        {/if}
                    </aside>
                    <ol>
                        {#each resultsToShow as result}
                            {@const component = getSearchResultComponent(result)}
                            <li><svelte:component this={component} {result} /></li>
                        {/each}
                        <div use:observeIntersection on:intersecting={loadMore} />
                    </ol>
                </div>
                <aside class="sidebar">
                    <h4>Filters</h4>
                    <Section
                        id={SearchSidebarSectionID.SEARCH_TYPES}
                        items={searchTypes}
                        title="Search types"
                        on:click={updateQuery}
                    />
                    {#if langFilters.length > 1}
                        <Section
                            id={SearchSidebarSectionID.LANGUAGES}
                            items={langFilters}
                            title="Languages"
                            on:click={updateQuery}
                        />
                    {/if}
                </aside>
            {/if}
        </div>
    </div>
</section>

<style lang="scss">
    .search {
        border-bottom: 1px solid var(--border-color);
        align-self: stretch;
        padding: 0.5rem 1rem;
    }

    section {
        flex: 1;
        display: flex;
        align-items: center;
        flex-direction: column;
        overflow: hidden;
    }

    .results {
        flex: 1;
        align-self: stretch;
        overflow: auto;

        .scroll-container {
            padding: 1rem;
            display: flex;

            .spinner {
                flex: 1;
                display: flex;
                justify-content: center;
            }
        }
    }

    ol {
        padding: 0;
        margin: 0;
        list-style: none;

        li {
            margin-bottom: 1rem;
        }
    }

    .main {
        flex: 1 1 auto;
        min-width: 0;
    }

    .sidebar {
        margin-left: 1rem;
        position: sticky;
        top: 1rem;
        align-self: flex-start;
        width: 15.5rem;
        flex-shrink: 0;
        background-color: var(--sidebar-bg);
        border: 1px solid var(--sidebar-border-color);
        padding: 0.75rem;
        border-radius: var(--border-radius);

        h4 {
            margin: -0.25rem -0.75rem 0.75rem;
            padding: 0 0.75rem 0.5rem;
            border-bottom: 1px solid var(--sidebar-border-color);
        }
    }
</style>
