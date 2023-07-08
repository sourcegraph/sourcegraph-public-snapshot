<svelte:options immutable />

<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { fetchFileRangeMatches } from '$lib/search/api/highlighting'
    import { getSymbolIconPath } from '$lib/search/symbolIcons'
    import { displayRepoName, splitPath, getFileMatchUrl, getRepositoryUrl, type SymbolMatch } from '$lib/shared'

    import CodeExcerpt from './CodeExcerpt.svelte'
    import SearchResult from './SearchResult.svelte'

    export let result: SymbolMatch

    $: repoName = result.repository
    $: repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    $: [fileBase, fileName] = splitPath(result.path)
    $: ranges = result.symbols.map(symbol => ({
        startLine: symbol.line - 1,
        endLine: symbol.line,
    }))

    async function fetchHighlightedSymbolMatchLineRanges(startLine: number, endLine: number) {
        const highlightedSymbols = await fetchFileRangeMatches({ result, ranges })
        return highlightedSymbols[
            result.symbols.findIndex(symbol => symbol.line - 1 === startLine && symbol.line === endLine)
        ]
    }
</script>

<SearchResult {result}>
    <div slot="title">
        <a href={repoAtRevisionURL}>{displayRepoName(repoName)}</a>
        <span aria-hidden={true}>›</span>
        <a href={getFileMatchUrl(result)}>
            {#if fileBase}{fileBase}/{/if}<strong>{fileName}</strong>
        </a>
    </div>
    {#each result.symbols as symbol}
        <div class="result">
            <div class="symbol-icon--kind-{symbol.kind.toLowerCase()}">
                <Icon svgPath={getSymbolIconPath(symbol.kind)} inline />
            </div>
            <CodeExcerpt
                startLine={symbol.line - 1}
                endLine={symbol.line}
                fetchHighlightedFileRangeLines={fetchHighlightedSymbolMatchLineRanges}
                --background-color="transparent"
            />
        </div>
    {/each}
</SearchResult>

<style lang="scss">
    @import '@sourcegraph/shared/src/symbols/SymbolIcon.module.scss';

    .result {
        margin-bottom: 0.5rem;
        display: flex;
        align-items: center;
        width: 100%;
        background-color: var(--color-bg-2);
        padding: 0.25rem;
        border-radius: var(--border-radius);
    }
</style>
