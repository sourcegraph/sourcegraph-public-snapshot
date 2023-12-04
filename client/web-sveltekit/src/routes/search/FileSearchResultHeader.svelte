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

    export let result: ContentMatch | PathMatch | SymbolMatch

    $: repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    $: fileURL = getFileMatchUrl(result)
    $: repoName = displayRepoName(result.repository)
    $: [fileBase, fileName] = splitPath(result.path)

    let matches: [number, number][] = []
    $: if (result.type !== 'symbol' && result.pathMatches) {
        matches = result.pathMatches.map((match): [number, number] => [match.start.column, match.end.column])
    }
</script>

<a href={repoAtRevisionURL}>{repoName}</a>
<span aria-hidden={true}>&nbsp;›&nbsp;</span>
<!-- #key is needed here to recreate the link because use:highlightNode changes the DOM -->
{#key result}
    <a href={fileURL} use:highlightRanges={{ ranges: matches }}>
        {#if fileBase}{fileBase}/{/if}<strong>{fileName}</strong>
    </a>
{/key}
