<script lang="ts" context="module">
    type HistoryStore = InfinityQueryStore<HistoryPanel_HistoryConnection['nodes'], { afterCursor: string | null }>
    export interface Capture {
        history: ReturnType<HistoryStore['capture']>
        scroller?: ScrollerCapture
    }
</script>

<script lang="ts">
    import { page } from '$app/stores'
    import Avatar from '$lib/Avatar.svelte'
    import { SourcegraphURL } from '$lib/common'
    import { scrollIntoViewOnMount } from '$lib/dom'
    import type { InfinityQueryStore } from '$lib/graphql'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import { replaceRevisionInURL } from '$lib/shared'
    import Timestamp from '$lib/Timestamp.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Alert, Badge } from '$lib/wildcard'

    import type { HistoryPanel_HistoryConnection } from './HistoryPanel.gql'

    export let history: HistoryStore
    export let enableInlineDiff: boolean = false
    export let enableViewAtCommit: boolean = false

    export function capture(): Capture {
        return {
            history: history.capture(),
            scroller: scroller?.capture(),
        }
    }

    export async function restore(data: Capture) {
        await history.restore(data.history)

        // If the selected revision is not in the set of currently loaded commits, load more
        if (selectedRev) {
            await history.fetchWhile(data => !data.find(commit => selectedRev?.startsWith(commit.abbreviatedOID)))
        }

        if (data.scroller) {
            // restore might be called when the history panel is closed
            // in which case scroller doesn't exist
            scroller?.restore(data.scroller)
        }
    }

    let scroller: Scroller

    $: selectedRev = $page.url?.searchParams.get('rev')
    $: diffEnabled = $page.url?.searchParams.has('diff')
    $: closeURL = SourcegraphURL.from($page.url).deleteSearchParameter('rev', 'diff').toString()
</script>

<Scroller bind:this={scroller} margin={200} on:more={history.fetchMore}>
    {#if $history.data}
        <table>
            {#each $history.data as commit (commit.id)}
                {@const selected = commit.abbreviatedOID === selectedRev || commit.oid === selectedRev}
                <tr class:selected use:scrollIntoViewOnMount={selected}>
                    <td>
                        <Badge variant="link"><a href={commit.canonicalURL}>{commit.abbreviatedOID}</a></Badge>
                    </td>
                    <td class="subject">
                        {#if enableInlineDiff}
                            <a href={selected ? closeURL : `?rev=${commit.oid}&diff=1`}>{commit.subject}</a>
                        {:else}
                            {commit.subject}
                        {/if}
                    </td>
                    <td>
                        <Avatar avatar={commit.author.person} />&nbsp;
                        {commit.author.person.displayName}
                    </td>
                    <td><Timestamp date={new Date(commit.author.date)} strict /></td>
                    {#if enableViewAtCommit}
                        <td>
                            <Tooltip tooltip={selected && !diffEnabled ? 'Close commit' : 'View at commit'}>
                                <a href={selected && !diffEnabled ? closeURL : `?rev=${commit.oid}`}
                                    ><Icon icon={ILucideFileText} inline aria-hidden /></a
                                >
                            </Tooltip>
                        </td>
                    {/if}
                    <td>
                        <Tooltip tooltip="Browse files at commit">
                            <a
                                href={replaceRevisionInURL(
                                    SourcegraphURL.from($page.url).deleteSearchParameter('rev', 'diff').toString(),
                                    commit.oid
                                )}><Icon icon={ILucideFolderGit} inline aria-hidden /></a
                            >
                        </Tooltip>
                    </td>
                </tr>
            {/each}
        </table>
    {/if}
    {#if $history.fetching}
        <div class="info">
            <LoadingSpinner />
        </div>
    {:else if $history.error}
        <div class="info">
            <Alert variant="danger">Unable to load history: {$history.error.message}</Alert>
        </div>
    {/if}
</Scroller>

<style lang="scss">
    .info {
        padding: 0.5rem 1rem;
    }

    table {
        width: 100%;
        max-width: 100%;
    }

    td {
        padding: 0.5rem 1rem;
        white-space: nowrap;

        :global([data-avatar]) {
            vertical-align: middle;
        }

        &.subject {
            white-space: normal;
        }
    }

    tr {
        --icon-color: var(--header-icon-color);
        border-bottom: 1px solid var(--border-color);

        &.selected {
            --icon-color: currentColor;

            color: var(--light-text);
            background-color: var(--primary);

            a {
                color: inherit;
            }
        }
    }
</style>
