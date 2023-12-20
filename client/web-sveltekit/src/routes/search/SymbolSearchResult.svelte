<svelte:options immutable />

<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { fetchFileRangeMatches } from '$lib/search/api/highlighting'
    import { getSymbolIconPath } from '$lib/search/symbolIcons'
    import type { SymbolMatch } from '$lib/shared'
    import FileSearchResultHeader from './FileSearchResultHeader.svelte'
    import { observeIntersection } from '$lib/intersection-observer'

    import CodeExcerpt from '$lib/search/CodeExcerpt.svelte'
    import CodeHostIcon from './CodeHostIcon.svelte'
    import RepoStars from './RepoStars.svelte'
    import SearchResult from './SearchResult.svelte'

    export let result: SymbolMatch

    $: ranges = result.symbols.map(symbol => ({
        startLine: symbol.line - 1,
        endLine: symbol.line,
    }))

    let hasBeenVisible = false
    let highlightedHTMLRows: string[][] = undefined
    async function onIntersection(event: { detail: boolean }) {
        if (hasBeenVisible) {
            return
        }
        hasBeenVisible = true
        highlightedHTMLRows = await fetchFileRangeMatches({ result, ranges: ranges })
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
    <svelte:fragment slot="body">
        <div use:observeIntersection on:intersecting={onIntersection}>
            {#each result.symbols as symbol, index}
                <a href={symbol.url}>
                    <div class="result">
                        <div class="symbol-icon--kind-{symbol.kind.toLowerCase()}">
                            <Icon svgPath={getSymbolIconPath(symbol.kind)} inline />
                        </div>
                        <CodeExcerpt
                            startLine={symbol.line - 1}
                            plaintextLines={['']}
                            highlightedHTMLRows={highlightedHTMLRows?.[index]}
                            --background-color="transparent"
                        />
                    </div>
                </a>
            {/each}
        </div>
    </svelte:fragment>
</SearchResult>

<style lang="scss">
    @import '@sourcegraph/shared/src/symbols/SymbolIcon.module.scss';

    .result {
        display: flex;
        align-items: center;
        width: 100%;
        background-color: var(--code-bg);
        padding: 0.25rem;
        border-bottom: 1px solid var(--border-color);
    }

    a {
        text-decoration: none;
    }
</style>
