<script context="module">
    export const focusContainerClass = 'copy-button-focus-container'
</script>

<script lang="ts">
    import { mdiContentCopy } from '@mdi/js'
    import copy from 'copy-to-clipboard'

    import { highlightRanges } from '$lib/dom'
    import Icon from '$lib/Icon.svelte'
    import {
        displayRepoName,
        splitPath,
        getFileMatchUrl,
        getRepositoryUrl,
        type ContentMatch,
        type PathMatch,
        type SymbolMatch,
    } from '$lib/shared'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Button } from '$lib/wildcard'

    export let result: ContentMatch | PathMatch | SymbolMatch

    $: repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    $: fileURL = getFileMatchUrl(result)
    $: repoName = displayRepoName(result.repository)
    $: [fileBase, fileName] = splitPath(result.path)

    $: matches =
        result.type !== 'symbol' && result.pathMatches
            ? result.pathMatches.map((match): [number, number] => [match.start.column, match.end.column])
            : []

    let recentlyCopied = false
    function handleCopyPath(): void {
        copy(result.path)
        recentlyCopied = true
        setTimeout(() => {
            recentlyCopied = false
        }, 1000)
    }

    $: tooltip = recentlyCopied ? 'Copied!' : 'Copy path to clipboard'
</script>

<a href={repoAtRevisionURL}>{repoName}</a>
<span aria-hidden={true}>&nbsp;›&nbsp;</span>
<!-- #key is needed here to recreate the link because use:highlightNode changes the DOM -->
<span class="root">
    {#key result}
        <a href={fileURL} use:highlightRanges={{ ranges: matches }}>
            {#if fileBase}{fileBase}/{/if}<strong>{fileName}</strong>
        </a>
        <div class="copy-button">
            <Tooltip {tooltip} placement="bottom">
                <Button on:click={() => handleCopyPath()} variant="icon" size="sm" aria-label="Copy path to clipboard">
                    <Icon inline svgPath={mdiContentCopy} aria-hidden />
                </Button>
            </Tooltip>
        </div>
    {/key}
</span>

<style lang="scss">
    .root {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
    .copy-button {
        visibility: hidden;
        --color: var(--icon-color);
        &:hover {
            --color: var(--body-color);
        }
    }

    :global(.copy-button-focus-container:hover),
    :global(.copy-button-focus-container:focus-within) {
        .copy-button {
            visibility: visible;
        }
    }
</style>
