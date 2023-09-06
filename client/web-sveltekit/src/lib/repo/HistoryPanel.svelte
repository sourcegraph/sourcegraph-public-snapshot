<script lang="ts" context="module">
    export interface Capture {
        history: HistoryResult | null
        scroller?: ScrollerCapture
    }
</script>

<script lang="ts">
    import { mdiClose } from '@mdi/js'
    import { tick } from 'svelte'

    import { page } from '$app/stores'
    import { scrollIntoView } from '$lib/actions'
    import type { HistoryResult } from '$lib/graphql-operations'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { createHistoryPanelStore } from '$lib/repo/stores'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import UserAvatar from '$lib/UserAvatar.svelte'
    import Timestamp from '$lib/Timestamp.svelte'

    export let history: Promise<HistoryResult>
    export let fetchMoreHandler: (pageInfo: HistoryResult['pageInfo']) => Promise<HistoryResult>

    export function capture(): Capture {
        return {
            history: historyStore.capture(),
            scroller: scroller?.capture(),
        }
    }

    export async function restore(data: Capture) {
        historyStore.restore(data.history)

        if (data.scroller) {
            // Wait until DOM was update before updating the scroll position
            await tick()
            // restore might be called when the history panel is closed
            // in which case scroller doesn't exist
            scroller?.restore(data.scroller)
        }
    }

    function getClearURL(): string {
        const url = new URL($page.url ?? window.location)
        url.searchParams.delete('rev')
        return url.href
    }

    const historyStore = createHistoryPanelStore(history)
    $: loading = $historyStore.loading
    $: resolvedCommits = $historyStore.history?.nodes ?? []

    function loadMore() {
        historyStore.loadMore(fetchMoreHandler)
    }

    // If the selected revision is not in the set of currentl loaded commits, load more
    $: if (
        selectedRev &&
        resolvedCommits.length > 0 &&
        !resolvedCommits.some(commit => commit.abbreviatedOID === selectedRev)
    ) {
        loadMore()
    }

    $: selectedRev = $page.url?.searchParams.get('rev')
    $: clearURL = getClearURL()

    let scroller: Scroller
</script>

<Scroller bind:this={scroller} margin={200} on:more={loadMore}>
    <table>
        {#each resolvedCommits as commit (commit.id)}
            {@const selected = commit.abbreviatedOID === selectedRev}
            <tr class:selected use:scrollIntoView={selected}>
                <td>
                    <UserAvatar user={commit.author.person} />&nbsp;
                    {commit.author.person.displayName}
                </td>
                <td class="subject">
                    {#if $page.params?.path}
                        <a href="?rev={commit.abbreviatedOID}">{commit.subject}</a>
                    {:else}
                        {commit.subject}
                    {/if}
                </td>
                <td><Timestamp date={new Date(commit.author.date)} /></td>
                <td><a href={commit.canonicalURL}>{commit.abbreviatedOID}</a></td>
                <td>
                    {#if selected}
                        <Tooltip tooltip="Hide comparison">
                            <a href={clearURL}><Icon svgPath={mdiClose} inline /></a>
                        </Tooltip>
                    {/if}
                </td>
            </tr>
        {/each}
    </table>
    {#if loading}
        <LoadingSpinner />
    {/if}
</Scroller>

<style lang="scss">
    table {
        width: 100%;
        max-width: 100%;
        overflow-y: auto;
    }

    td {
        padding: 0.25rem;
        white-space: nowrap;

        &.subject {
            white-space: normal;
        }
    }

    tr {
        border-bottom: 1px solid var(--border-color);

        &.selected {
            background-color: var(--color-bg-2);
        }
    }
</style>
