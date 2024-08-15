<script lang="ts" context="module">
    type HistoryPagination = Pagination<HistoryPanel_HistoryConnection['nodes'], { afterCursor: string | null }>
    export interface Capture {
        historyPagination: ReturnType<HistoryPagination['capture']>
        scroller?: ScrollerCapture
    }
</script>

<script lang="ts">
    import { page } from '$app/stores'
    import Avatar from '$lib/Avatar.svelte'
    import { SourcegraphURL } from '$lib/common'
    import { scrollIntoViewOnMount } from '$lib/dom'
    import type { Pagination } from '$lib/graphql'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import { replaceRevisionInURL } from '$lib/shared'
    import Timestamp from '$lib/Timestamp.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Alert, Badge } from '$lib/wildcard'

    import type { HistoryPanel_HistoryConnection } from './HistoryPanel.gql'

    export let historyPagination: HistoryPagination
    export let enableInlineDiff: boolean = false
    export let enableViewAtCommit: boolean = false

    export function capture(): Capture {
        return {
            historyPagination: historyPagination.capture(),
            scroller: scroller?.capture(),
        }
    }

    export async function restore(data: Capture) {
        await historyPagination.restore(data.historyPagination)

        // If the selected revision is not in the set of currently loaded commits, load more
        if (selectedRev) {
            await historyPagination.fetchNextWhile(data => !data.find(commit => selectedRev?.startsWith(commit.abbreviatedOID)))
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

<Scroller bind:this={scroller} margin={200} on:more={historyPagination.fetchMore}>
    {#if $historyPagination.data}
        <table>
            <tbody>
                {#each $historyPagination.data as commit (commit.id)}
                    {@const selected = commit.abbreviatedOID === selectedRev || commit.oid === selectedRev}
                    <tr class:selected use:scrollIntoViewOnMount={selected}>
                        <td class="revision">
                            <Badge variant="link"><a href={commit.canonicalURL}>{commit.abbreviatedOID}</a></Badge>
                        </td>
                        <td class="subject">
                            {#if enableInlineDiff}
                                <a href={selected ? closeURL : `?rev=${commit.oid}&diff=1`}>{commit.subject}</a>
                            {:else}
                                {commit.subject}
                            {/if}
                        </td>
                        <td class="author">
                            <Avatar avatar={commit.author.person} />&nbsp;
                            {commit.author.person.displayName}
                        </td>
                        <td class="timestamp"><Timestamp date={new Date(commit.author.date)} strict /></td>
                        <td class="actions">
                            {#if enableViewAtCommit}
                                <Tooltip tooltip={selected && !diffEnabled ? 'Close commit' : 'View at commit'}>
                                    <a href={selected && !diffEnabled ? closeURL : `?rev=${commit.oid}`}
                                        ><Icon icon={ILucideFileText} inline aria-hidden /></a
                                    >
                                </Tooltip>
                            {/if}
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
            </tbody>
        </table>
    {/if}
    {#if $historyPagination.loading}
        <div class="info">
            <LoadingSpinner />
        </div>
    {:else if $historyPagination.error}
        <div class="info">
            <Alert variant="danger">Unable to load history: {$historyPagination.error.message}</Alert>
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
        display: grid;
        grid-template-columns: [revision] auto [subject] 3fr [author] 1fr [timestamp] auto [actions] auto;
    }

    tbody,
    tr {
        display: grid;
        grid-column: 1/-1;
        grid-template-columns: subgrid;
    }

    td {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        padding: 0.5rem;

        :global([data-avatar]) {
            vertical-align: middle;
        }
    }

    .timestamp {
        grid-area: timestamp;
        text-align: right;
        color: var(--text-muted);
    }

    .revision {
        grid-area: revision;
    }

    .subject {
        grid-area: subject;
    }

    .author {
        grid-area: author;
    }

    .actions {
        display: flex;
        gap: 1rem;
        padding: 0.5rem 1rem;
        grid-area: actions;
    }

    tr {
        border-bottom: 1px solid var(--border-color);
        font-size: var(--font-size-small);

        &.selected {
            --icon-color: currentColor;

            color: var(--light-text);
            background-color: var(--primary);

            a,
            .timestamp {
                color: inherit;
            }
        }
    }

    @media (--mobile) {
        table {
            grid-template-columns: 1fr auto auto auto;
        }
        tr {
            grid-template-areas: 'subject subject subject revision' 'author author timestamp actions';
        }

        .subject {
            white-space: normal;
        }

        .author,
        .timestamp,
        .actions {
            align-self: center;
        }

        .actions a {
            display: inline-block;
            padding: 0.5rem;
        }
    }
</style>
