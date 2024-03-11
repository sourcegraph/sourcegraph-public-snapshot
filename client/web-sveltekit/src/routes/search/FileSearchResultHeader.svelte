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

    import { getSearchResultsContext } from './searchResultsContext'

    export let result: ContentMatch | PathMatch | SymbolMatch

    $: repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    $: fileURL = getFileMatchUrl(result)
    $: repoName = displayRepoName(result.repository)
    $: [fileBase, fileName] = splitPath(result.path)

    $: matches =
        result.type !== 'symbol' && result.pathMatches
            ? result.pathMatches.map((match): [number, number] => [match.start.column, match.end.column])
            : []

    const searchResultContext = getSearchResultsContext()
    function handlePreview() {
        searchResultContext.setPreview({
            repoName: result.repository,
            commitID: result.commit,
            filePath: result.path,
            matchedRanges: [], // TODO
        })
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
<button on:click={handlePreview}>Preview</button>
