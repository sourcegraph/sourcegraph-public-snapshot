<svelte:options immutable />

<script lang="ts" context="module">
    const BY_LINE_RANKING = 'by-line-number'
    const DEFAULT_CONTEXT_LINES = 1
    const DEFAULT_EXPANDED_MATCHES = 3
</script>

<script lang="ts">
    import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
    import { observeIntersection } from '$lib/intersection-observer'

    import {
        addLineRangeQueryParameter,
        formatSearchParameters,
        pluralize,
        toPositionOrRangeQueryParameter,
    } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import { getFileMatchUrl, type ContentMatch, rankByLine, rankPassthrough } from '$lib/shared'

    import SearchResult from './SearchResult.svelte'
    import { getSearchResultsContext } from './searchResultsContext'
    import CodeHostIcon from './CodeHostIcon.svelte'
    import RepoStars from './RepoStars.svelte'
    import { settings } from '$lib/stores'
    import { rankContentMatch } from '$lib/search/results'
    import FileSearchResultHeader from './FileSearchResultHeader.svelte'
    import { fetchFileRangeMatches } from '$lib/search/api/highlighting'
    import CodeExcerpt from '$lib/search/CodeExcerpt.svelte'

    export let result: ContentMatch

    $: contextLines = $settings?.['search.contextLines'] ?? DEFAULT_CONTEXT_LINES
    $: ranking =
        $settings?.experimentalFeatures?.clientSearchResultRanking === BY_LINE_RANKING
            ? rankByLine
            : rankPassthrough
    $: ({ expandedMatchGroups, collapsedMatchGroups, hiddenMatchesCount } = rankContentMatch(
        result,
        ranking,
        DEFAULT_EXPANDED_MATCHES,
        contextLines
    ))
    $: collapsible = hiddenMatchesCount > 0
    $: fileURL = getFileMatchUrl(result)

    const searchResultContext = getSearchResultsContext()
    let expanded: boolean = searchResultContext?.isExpanded(result)
    $: searchResultContext.setExpanded(result, expanded)
    $: expandButtonText = expanded
        ? 'Show less'
        : `Show ${hiddenMatchesCount} more ${pluralize('match', hiddenMatchesCount, 'matches')}`
    $: matchesToShow = expanded ? expandedMatchGroups : collapsedMatchGroups

    let root: HTMLElement
    let userInteracted = false
    $: if (!expanded && root && userInteracted) {
        setTimeout(() => {
            const reducedMotion = !window.matchMedia('(prefers-reduced-motion: no-preference)').matches
            root.scrollIntoView({ block: 'nearest', behavior: reducedMotion ? 'auto' : 'smooth' })
        }, 0)
    }

    function getMatchURL(startLine: number, endLine: number): string {
        const searchParams = formatSearchParameters(
            addLineRangeQueryParameter(
                // We don't want to preserve the 'q' query parameter.
                // We might have to adjust this if we want to preserve other query parameters.
                new URLSearchParams(),
                toPositionOrRangeQueryParameter({ range: { start: { line: startLine }, end: { line: endLine } } })
            )
        )
        return `${fileURL}?${searchParams}`
    }

    let hasBeenVisible = false
    let highlightedHTMLRows: string[][] = undefined
    async function onIntersection(event: { detail: boolean }) {
        if (hasBeenVisible) {
            return
        }
        hasBeenVisible = true
        const matchRanges = expandedMatchGroups.map(group => ({
            startLine: group.startLine,
            endLine: group.endLine,
        }))
        highlightedHTMLRows = await fetchFileRangeMatches({ result, ranges: matchRanges })
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

    <div bind:this={root} use:observeIntersection on:intersecting={onIntersection} class="matches">
        {#each matchesToShow as group, index}
            <div class="code">
                <a href={getMatchURL(group.startLine + 1, group.endLine)}>
                    <CodeExcerpt
                        startLine={group.startLine}
                        matches={group.matches}
                        plaintextLines={group.plaintextLines}
                        highlightedHTMLRows={highlightedHTMLRows?.[index]}
                    />
                </a>
            </div>
        {/each}
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
    </div>
</SearchResult>

<style lang="scss">
    button {
        width: 100%;
        text-align: left;
        border: none;
        padding: 0.25rem 0.5rem;
        background-color: var(--code-bg);
        color: var(--collapse-results-color);
        cursor: pointer;

        &.expanded {
            position: sticky;
            bottom: 0;
        }
    }

    .code {
        border-bottom: 1px solid var(--border-color);

        &:last-child {
            border-bottom: none;
        }

        a {
            text-decoration: none;
            color: inherit;
        }
    }
</style>
