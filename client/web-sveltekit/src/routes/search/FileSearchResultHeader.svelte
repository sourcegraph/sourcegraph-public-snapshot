<script lang="ts">
    import { highlightRanges } from '$lib/dom'
    import {
        displayRepoName,
        splitPath,
        getFileMatchUrl,
        getRepositoryUrl,
        type ContentMatch,
        type PathMatch,
        type SymbolMatch,
    } from '$lib/shared'

    import CopyPathButton from './CopyPathButton.svelte'

    export let result: ContentMatch | PathMatch | SymbolMatch

    $: repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    $: fileURL = getFileMatchUrl(result)
    $: repoName = displayRepoName(result.repository)
    $: [fileBase, fileName] = splitPath(result.path)
    $: rev = result.branches?.[0]

    $: matches =
        result.type !== 'symbol' && result.pathMatches
            ? result.pathMatches.map((match): [number, number] => [match.start.column, match.end.column])
            : []
</script>

<a class="repo-link" href={repoAtRevisionURL}>
    {repoName}
    {#if rev}
        <span class="rev-tag">@ <small class="rev">{rev}</small></span>
    {/if}
</a>
<span class="interpunct" aria-hidden={true}>⋅</span>
<!-- #key is needed here to recreate the link because use:highlightNode changes the DOM -->
<span class="root">
    {#key result}
        <a href={fileURL} use:highlightRanges={{ ranges: matches }}>
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
        a {
            color: var(--text-body);
        }
    }

    .interpunct {
        margin: 0 0.5rem;
        color: var(--text-disabled);
    }

    .repo-link {
        color: var(--text-body);
    }

    .file-name {
        font-weight: 600;
    }

    .rev {
        &-tag {
            color: var(--text-muted);
        }
        background-color: var(--color-bg-2);
        padding: 0.25rem;
    }
</style>
