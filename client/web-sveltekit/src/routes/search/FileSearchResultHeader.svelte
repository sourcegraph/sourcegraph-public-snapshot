<script lang="ts">
    import { highlightRanges } from '$lib/dom'
    import { splitPath, getFileMatchUrl, type ContentMatch, type PathMatch, type SymbolMatch } from '$lib/shared'

    import CopyPathButton from './CopyPathButton.svelte'
    import RepoRev from './RepoRev.svelte'

    export let result: ContentMatch | PathMatch | SymbolMatch

    $: fileURL = getFileMatchUrl(result)
    $: [fileBase, fileName] = splitPath(result.path)
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
            {#if fileBase}{fileBase}/{/if}<span class="file-name">{fileName}</span>
        </a>
    {/key}
    <CopyPathButton path={result.path} />
</span>

<style lang="scss">
    .root {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-family: var(--code-font-family);
        font-size: var(--code-font-size);
    }

    .interpunct {
        margin: 0 0.5rem;
        color: var(--text-muted);
    }

    .path {
        color: var(--text-body);
        .file-name {
            font-weight: 600;
        }
    }
</style>
