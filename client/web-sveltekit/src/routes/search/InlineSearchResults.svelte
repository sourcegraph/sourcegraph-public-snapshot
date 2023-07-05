<script context="module" lang="ts">
    interface Cache {
        count: number
        expanded: Set<SearchMatch>
    }
    const cache = new Map<string, Cache>()

    export interface Context {
        isExpanded(match: SearchMatch): boolean
        setExpanded(match: SearchMatch, expanded: boolean): void
    }

    const DEFAULT_INITIAL_ITEMS_TO_SHOW = 15
    const INCREMENTAL_ITEMS_TO_SHOW = 10
</script>

<script lang="ts">
    import type { Observable } from 'rxjs'
    import { setContext } from 'svelte'

    import { beforeNavigate } from '$app/navigation'
    import { preserveScrollPosition } from '$lib/app'
    import { observeIntersection } from '$lib/intersection-observer'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import SearchBox from '$lib/search/SearchBox.svelte'
    import { searchTypes } from '$lib/search/sidebar'
    import { submitSearch, type QueryStateStore } from '$lib/search/state'
    import type { SidebarFilter } from '$lib/search/utils'
    import { type AggregateStreamingSearchResults, type SearchMatch, type ContentMatch, getRevision} from '$lib/shared'

    import StreamingProgress from './StreamingProgress.svelte'
    import InlineFileSearchResult from './InlineFileSearchResult.svelte'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import { createBlobStore } from '$lib/blob/ui/stores'
    import Divider, { getDividerStore } from '$lib/Divider.svelte'
    import { mapResultToCodeMirrorDecorations } from '$lib/search/ui/filepreview'
    import { EditorView } from '@codemirror/view'
    import type { Extension } from '@codemirror/state'

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
    setContext<Context>('search-results', {
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
    })
    beforeNavigate(() => {
        cache.set(queryFromURL, { count, expanded: expandedSet })
    })

    function loadMore(event: { detail: boolean }) {
        if (event.detail) {
            count += INCREMENTAL_ITEMS_TO_SHOW
        }
    }

    let blobStore = createBlobStore()
    let selectedResult: ContentMatch|null = null
    let selectedLine: number|null = null

    $: if (!results) {
        blobStore.fetch(null)
    }
    $: {
        blobStore.fetch(selectedResult ? {filePath: selectedResult.path, repoName: selectedResult.repository, revision: getRevision(selectedResult.branches, selectedResult.commit)} : null)
    }
    $: showFilePreview = $blobStore.loading || $blobStore.blob

    function selectResult({detail: selected}: CustomEvent<{result: ContentMatch, line: number}>) {
        if (selectedResult !== selected.result) {
            selectedResult = selected.result
        }
        selectedLine = selected.line
    }

    let dividerPosition = getDividerStore('file-preview', 0.5)
    $: minResultsWidth = `${$dividerPosition * 100}%`
    $: maxResultsWidth = selectedResult && !loading ? minResultsWidth : undefined
</script>

<section>
    <div class="search">
        <SearchBox bind:this={searchInput} {queryState} />
    </div>

    <div class="results" bind:this={resultContainer}>
        <div class="scroll-container" style:min-width={minResultsWidth} style:max-width={maxResultsWidth}>
            {#if !$stream || loading}
                <div class="spinner">
                    <LoadingSpinner />
                </div>
            {:else if !loading && resultsToShow}
                <div class="main">
                    <ol>
                        {#each resultsToShow as result}
                            <li>
                                {#if result.type === 'content'}
                                    <InlineFileSearchResult {result} on:select={selectResult} selectedLine={result === selectedResult ? selectedLine : null}/>
                                <!--
                                {:else if result.type === 'repo'}
                                    <RepoSearchResult {result} />
                                {:else if result.type === 'symbol'}
                                    <SymbolSearchResult {result} />
                                -->
                                {/if}
                            </li>
                        {/each}
                        <div use:observeIntersection on:intersecting={loadMore} />
                    </ol>
                </div>
            {/if}
        </div>
        {#if showFilePreview}
            <Divider id="file-preview" />
            <div class="file-preview">
            {#if $blobStore.loading}
                <LoadingSpinner />
            {:else if $blobStore.blob}
                    <CodeMirrorBlob
                        blob={$blobStore.blob}
                        highlights={$blobStore.highlights ?? ''}
                        focusedLine={selectedLine ?? undefined}
                        wrapLines={true}
                        extension={selectedResult ? EditorView.decorations.of(mapResultToCodeMirrorDecorations(selectedResult)) : []}
                    />
            {/if}
            </div>
        {/if}
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
        display: flex;
        flex: 1;
        align-self: stretch;
        overflow: hidden;

        .scroll-container {
            margin-left: 1rem;
            overflow: auto;
            flex: 1;

            .spinner {
                flex: 1;
                display: flex;
                justify-content: center;
            }
        }

        .file-preview {
            flex: 1;
            background-color: var(--code-bg);
            flex-basis: 0px;
            overflow: auto;
            border-left: 1px solid var(--border-color);
        }
    }

    ol {
        padding: 0;
        margin: 0;
        list-style: none;
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
