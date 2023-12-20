<svelte:options immutable />

<script context="module" lang="ts">
    interface ResultStateCache {
        count: number
        expanded: Set<SearchMatch>
    }
    const cache = new Map<string, ResultStateCache>()

    const DEFAULT_INITIAL_ITEMS_TO_SHOW = 15
    const INCREMENTAL_ITEMS_TO_SHOW = 10
</script>

<script lang="ts">
    import type { Observable } from 'rxjs'
    import { tick } from 'svelte'

    import { beforeNavigate } from '$app/navigation'
    import { preserveScrollPosition } from '$lib/app'
    import { observeIntersection } from '$lib/intersection-observer'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import SearchInput from '$lib/search/input/SearchInput.svelte'
    import { resultTypeFilter } from '$lib/search/sidebar'
    import { submitSearch, type QueryStateStore, getQueryURL } from '$lib/search/state'
    import { groupFilters } from '$lib/search/utils'
    import { type AggregateStreamingSearchResults, displayRepoName, type SearchMatch } from '$lib/shared'

    import Section from './SidebarSection.svelte'
    import StreamingProgress from './StreamingProgress.svelte'
    import { getSearchResultComponent } from './searchResultFactory'
    import { setSearchResultsContext } from './searchResultsContext'
    import Separator, { getSeparatorPosition } from '$lib/Separator.svelte'
    import Icon from '$lib/Icon.svelte'
    import { mdiBookOpenVariant, mdiCloseOctagonOutline } from '@mdi/js'
    import CodeHostIcon from './CodeHostIcon.svelte'

    export let stream: Observable<AggregateStreamingSearchResults | undefined>
    export let queryFromURL: string
    export let queryFilters: string
    export let queryState: QueryStateStore

    let resultContainer: HTMLElement | null = null
    let searchInput: SearchInput

    const sidebarSize = getSeparatorPosition('search-results-sidebar', 0.2)

    $: sidebarWidth = `max(100px, min(50%, ${$sidebarSize * 100}%))`
    $: progress = $stream?.progress
    // NOTE: done is present but apparently not officially exposed. However
    // $stream.state is always "loading". Need to look into this.
    $: loading = !progress?.done
    $: results = $stream?.results
    $: filters = groupFilters($stream?.filters)
    $: hasFilters = filters.lang.length > 0 || filters.repo.length > 0 || filters.file.length > 0

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

<svelte:head>
    <title>{queryFromURL} - Sourcegraph</title>
</svelte:head>

<div class="search">
    <SearchInput bind:this={searchInput} {queryState} showSmartSearchButton />
</div>

<div class="search-results">
    <aside class="sidebar" style:width={sidebarWidth}>
        <div class="section">
            <!-- TODO: a11y -->
            <ul>
                {#each resultTypeFilter as filter}
                    <li class:selected={filter.isSelected(queryFromURL)}>
                        <a
                            href={getQueryURL({
                                searchMode: $queryState.searchMode,
                                patternType: $queryState.patternType,
                                caseSensitive: $queryState.caseSensitive,
                                searchContext: $queryState.searchContext,
                                query: filter.getQuery($queryState.query),
                            })}
                        >
                            <Icon svgPath={filter.icon} inline aria-hidden="true" />
                            {filter.label}
                        </a>
                    </li>
                {/each}
            </ul>
        </div>
        {#if hasFilters}
            <div class="section">
                <h4>Filter results</h4>
                {#if filters.lang.length > 0}
                    <Section items={filters.lang} title="By languages" {queryFilters} />
                {/if}
                {#if filters.repo.length > 0}
                    <Section items={filters.repo} title="By repositories" {queryFilters}>
                        <svelte:fragment slot="label" let:label>
                            <CodeHostIcon repository={label} />
                            {displayRepoName(label)}
                        </svelte:fragment>
                    </Section>
                {/if}
                {#if filters.file.length > 0}
                    <Section items={filters.file} title="By paths" {queryFilters} />
                {/if}
            </div>
        {/if}
        <a class="section help" href="/help/code_search/reference/queries" target="_blank">
            <span class="icon">
                <Icon svgPath={mdiBookOpenVariant} inline />
            </span>
            <div>
                <h4>Need more advanced filters?</h4>
                <span>Explore the query syntax docs</span>
            </div>
        </a>
    </aside>
    <Separator currentPosition={sidebarSize} />
    <div class="results" bind:this={resultContainer}>
        <aside class="actions">
            {#if loading}
                <div>
                    <LoadingSpinner inline />
                </div>
            {/if}
            {#if progress}
                <StreamingProgress {progress} on:submit={onResubmitQuery} />
            {/if}
        </aside>
        {#if resultsToShow}
            <ol>
                {#each resultsToShow as result}
                    {@const component = getSearchResultComponent(result)}
                    <li><svelte:component this={component} {result} /></li>
                {/each}
                <div use:observeIntersection on:intersecting={loadMore} />
            </ol>
            {#if resultsToShow.length === 0 && !loading}
                <div class="no-result">
                    <Icon svgPath={mdiCloseOctagonOutline} />
                    <p>No results found</p>
                </div>
            {/if}
        {/if}
    </div>
</div>

<style lang="scss">
    .search {
        border-bottom: 1px solid var(--border-color);
        align-self: stretch;
        padding: 0.25rem;
    }

    .search-results {
        display: flex;
        flex: 1;
        overflow: hidden;
    }

    .sidebar {
        flex: 0 0 auto;
        background-color: var(--sidebar-bg);
        overflow-y: auto;
        display: flex;
        flex-direction: column;

        h4 {
            font-weight: 600;
            white-space: nowrap;
            margin-bottom: 1rem;
        }

        .section {
            padding: 1rem;
            border-top: 1px solid var(--border-color);

            &:first-child {
                border-top: none;
            }

            &:last-child {
                margin-top: auto;
            }
        }

        ul {
            margin: 0;
            padding: 0;
            list-style: none;

            a {
                flex: 1;
                color: var(--sidebar-text-color);
                text-decoration: none;
                padding: 0.25rem 0.5rem;
                border-radius: var(--border-radius);
                // Controls icon color
                --color: var(--icon-color);

                &:hover {
                    background-color: var(--secondary-4);
                }
            }

            li {
                display: flex;
                white-space: nowrap;

                &.selected {
                    a {
                        background-color: var(--primary);
                        color: var(--primary-4);
                        --color: var(--primary-4);
                    }
                }
            }
        }
    }

    .results {
        flex: 1;
        overflow: auto;
        display: flex;
        flex-direction: column;

        .actions {
            border-bottom: 1px solid var(--border-color);
            padding: 0.5rem 0;
            padding-left: 0.25rem;
            display: flex;
            align-items: center;
            // Explictly set height to avoid jumping when loading spinner is
            // shown/hidden.
            height: 3rem;
            flex-shrink: 0;
        }

        ol {
            padding: 0;
            margin: 0;
            list-style: none;
        }

        .no-result {
            display: flex;
            flex-direction: column;
            align-items: center;
            margin: auto;
            color: var(--text-muted);
        }
    }

    .help {
        display: flex;
        align-items: center;
        gap: 0.5rem;

        text-decoration: none;
        color: var(--text-muted);
        font-size: 0.75rem;

        h4 {
            margin: 0;
        }

        .icon {
            flex-shrink: 0;
        }
    }
</style>
