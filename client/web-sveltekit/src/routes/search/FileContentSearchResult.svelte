<svelte:options immutable />

<script lang="ts" context="module">
    const BY_LINE_RANKING = 'by-line-number'
    const DEFAULT_CONTEXT_LINES = 1
    const MAX_LINE_MATCHES = 5
    const MAX_ZOEKT_RESULTS = 3
</script>

<script lang="ts">
    import { mdiChevronDown, mdiChevronUp } from '@mdi/js'

    import {
        addLineRangeQueryParameter,
        formatSearchParameters,
        pluralize,
        toPositionOrRangeQueryParameter,
    } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import { getFileMatchUrl, type ContentMatch, ZoektRanking, LineRanking } from '$lib/shared'

    import FileMatchChildren from './FileMatchChildren.svelte'
    import SearchResult from './SearchResult.svelte'
    import { getSearchResultsContext } from './SearchResults.svelte'
    import CodeHostIcon from './CodeHostIcon.svelte'
    import RepoStars from './RepoStars.svelte'
    import { settings } from '$lib/stores'
    import { rankContentMatch } from '$lib/search/results'
    import { goto } from '$app/navigation'
    import FileSearchResultHeader from './FileSearchResultHeader.svelte'

    export let result: ContentMatch

    $: contextLines = $settings?.['search.contextLines'] ?? DEFAULT_CONTEXT_LINES
    $: ranking =
        $settings?.experimentalFeatures?.clientSearchResultRanking === BY_LINE_RANKING
            ? new LineRanking(MAX_LINE_MATCHES)
            : new ZoektRanking(MAX_ZOEKT_RESULTS)
    $: ({ expandedMatchGroups, collapsedMatchGroups, collapsible, hiddenMatchesCount } = rankContentMatch(
        result,
        ranking,
        contextLines
    ))
    $: fileURL = getFileMatchUrl(result)

    const searchResultContext = getSearchResultsContext()
    let expanded: boolean = searchResultContext?.isExpanded(result)
    $: searchResultContext.setExpanded(result, expanded)
    $: expandButtonText = expanded
        ? 'Show less'
        : `Show ${hiddenMatchesCount} more ${pluralize('match', hiddenMatchesCount, 'matches')}`

    let root: HTMLElement
    let userInteracted = false
    $: if (!expanded && root && userInteracted) {
        setTimeout(() => {
            const reducedMotion = !window.matchMedia('(prefers-reduced-motion: no-preference)').matches
            root.scrollIntoView({ block: 'nearest', behavior: reducedMotion ? 'auto' : 'smooth' })
        }, 0)
    }

    function handleLineClick(event: MouseEvent) {
        const target = event.target as HTMLElement
        if (target.dataset.line) {
            const searchParams = formatSearchParameters(
                addLineRangeQueryParameter(
                    // We don't want to preserve the 'q' query parameter.
                    // We might have to adjust this if we want to preserver other query parameters.
                    new URLSearchParams(),
                    toPositionOrRangeQueryParameter({ position: { line: +target.dataset.line } })
                )
            )
            goto(`${fileURL}?${searchParams}`)
        }
    }
</script>

<SearchResult>
    <CodeHostIcon slot="icon" repository={result.repository} />
    <FileSearchResultHeader slot="title" {result} />
    <svelte:fragment slot="info">
        {#if result.repoStars}
            <RepoStars repoStars={result.repoStars} />
        {/if}
    </svelte:fragment>

    <div bind:this={root} class="matches" on:click={handleLineClick}>
        <FileMatchChildren {result} grouped={expanded ? expandedMatchGroups.grouped : collapsedMatchGroups.grouped} />
    </div>
    {#if collapsible}
        <button
            type="button"
            on:click={() => {
                expanded = !expanded
                userInteracted = true
            }}
            class:expanded
        >
            <Icon svgPath={expanded ? mdiChevronUp : mdiChevronDown} inline aria-hidden="true" />
            <span>{expandButtonText}</span>
        </button>
    {/if}
</SearchResult>

<style lang="scss">
    button {
        width: 100%;
        text-align: left;
        border: none;
        padding: 0.25rem 0.5rem;
        background-color: var(--border-color);
        border-radius: 0 0 var(--border-radius) var(--border-radius);
        color: var(--collapse-results-color);
        cursor: pointer;

        &.expanded {
            position: sticky;
            bottom: 0;
        }
    }

    .matches {
        // TODO: Evaluate whether (and how) these should/can be convertd to links
        :global(td[data-line]) {
            cursor: pointer;
            &:hover {
                text-decoration: underline;
            }
        }
    }
</style>
