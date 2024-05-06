<script lang="ts">
    import { highlightRanges } from '$lib/dom'
    import { getFileMatchUrl, type ContentMatch, type PathMatch, type SymbolMatch } from '$lib/shared'
    import CopyButton from '$lib/wildcard/CopyButton.svelte'

    import RepoRev from './RepoRev.svelte'

    export let result: ContentMatch | PathMatch | SymbolMatch

    $: fileURL = getFileMatchUrl(result)
    $: rev = result.branches?.[0]

    $: matches =
        result.type !== 'symbol' && result.pathMatches
            ? result.pathMatches.map((match): [number, number] => [match.start.column, match.end.column])
            : []
</script>

<RepoRev repoName={result.repository} {rev} />
<span class="interpunct">·</span>
<span class="root">
    {#key result}
        <a class="path" href={fileURL} use:highlightRanges={{ ranges: matches }}>
            {result.path}
        </a>
    {/key}
    <span data-visible-on-focus><CopyButton value={result.path} label="Copy path to clipboard" /></span>
</span>

<style lang="scss">
    .root {
        font-family: var(--code-font-family);
        font-size: var(--code-font-size);

        a,
        span {
            vertical-align: middle;
        }
    }

    .interpunct {
        margin: 0 0.5rem;
        color: var(--text-disabled);
    }

    .path {
        color: var(--text-body);
    }
</style>
