<svelte:options immutable />

<script lang="ts">
    import { mapContentMatchToMatchItems } from '$lib/search/api/results'
    import {
        displayRepoName,
        splitPath,
        getFileMatchUrl,
        getRepositoryUrl,
        type ContentMatch,
        type MatchItem,
    } from '$lib/shared'
    import { createEventDispatcher } from 'svelte'

    import SearchResult from './SearchResult.svelte'

    export let result: ContentMatch
    export let selectedLine: number|null = null


    const dispatch = createEventDispatcher<{select: {result: ContentMatch, line: number}}>()

    // TODO: Fix for mutliline matches
    function highlight(item: MatchItem): {content: string , match: boolean}[] {
        const result = []
        let index = 0
        for (const range of item.highlightRanges) {
            if (index < range.startCharacter) {
                result.push({content: item.content.slice(index, range.startCharacter), match: false})
            }
            result.push({content: item.content.slice(range.startCharacter, range.endCharacter), match: true})
            index = range.endCharacter
        }

        if (index < item.content.length) {
            result.push({content: item.content.slice(index), match: false})
        }

        return result
    }

    $: repoName = result.repository
    $: repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    $: [fileBase, fileName] = splitPath(result.path)
    $: items = mapContentMatchToMatchItems(result)
</script>

<SearchResult {result}>
    <div slot="title">
        <a href={repoAtRevisionURL}>{displayRepoName(repoName)}</a>
        <span aria-hidden={true}>›</span>
        <a href={getFileMatchUrl(result)}>
            {#if fileBase}{fileBase}/{/if}<strong>{fileName}</strong>
        </a>
    </div>
    <table>
        <tbody>
            {#each items as item}
                {@const matches = highlight(item)}
                {@const line = item.startLine + 1}
                <tr class:selected={line === selectedLine} on:click={() => dispatch("select", {result, line})}>
                    <td class="line" data-line={line} />
                    <td class="code">
                        <a href={getFileMatchUrl(result)} on:click={event => event.preventDefault()}>
                        {#each matches as match}
                            <span class:match-highlight={match.match}>{match.content}</span>
                        {/each  }
                        </a>
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</SearchResult>

<style lang="scss">
    .code {
        white-space: pre;
    }
    td.line::before {
        content: attr(data-line);
        color: var(--text-muted);
        padding-right: 1rem;
    }
    .line {
        text-align: right;
        width: 80px;
    }

    table {
        width: 100%;
        table-layout: fixed;
        border: 1px solid var(--border-color);
        background-color: var(--code-bg);
    }

    tr.selected {
        background-color: var(--color-bg-2);
    }

    tr:hover td {
        cursor: pointer;
        background-color: var(--color-bg-2);
    }

    tr:last-child td {
        border-bottom: none;
    }

    td {
        border-bottom: 1px solid var(--border-color);
        border-bottom: 1px solid var(--border-color);
    }

    a {
        color: var(--body-color);
        text-decoration: none;
    }
</style>
