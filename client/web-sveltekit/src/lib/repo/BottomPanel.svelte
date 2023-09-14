<script lang="ts" context="module">
    interface Capture {
        revision?: string
        history?: HistoryPanelCapture
        selectedTab: number | null
    }
</script>

<script lang="ts">
    import { tick } from 'svelte'

    import { goto } from '$app/navigation'
    import { page } from '$app/stores'
    import type { HistoryResult } from '$lib/graphql-operations'
    import { fetchRepoCommits } from '$lib/repo/api/commits'
    import TabPanel from '$lib/TabPanel.svelte'
    import Tabs from '$lib/Tabs.svelte'

    import HistoryPanel, { type Capture as HistoryPanelCapture } from './HistoryPanel.svelte'

    export let history: Promise<HistoryResult>

    export function capture(): Capture {
        return {
            revision: $page.data.resolvedRevision?.commitID,
            // History panel is not rendered when the bottom is closed
            history: historyPanel?.capture(),
            selectedTab,
        }
    }

    export async function restore(snapshot: Capture) {
        selectedTab = snapshot.selectedTab

        // We need to wait until the DOM was updated and the history panel is visible,
        // otherwise historyPanel won't be available yet.
        await tick()

        if (snapshot.history && snapshot.revision && snapshot.revision === $page.data.resolvedRevision?.commitID) {
            // Only restore history panel data if the stored revison matches the currently loaded revision
            historyPanel?.restore(snapshot.history)
        }
    }

    async function selectTab(event: { detail: number | null }) {
        if (event.detail === null) {
            const url = new URL($page.url)
            url.searchParams.delete('rev')
            await goto(url, { replaceState: true, keepFocus: true, noScroll: true })
        }
        selectedTab = event.detail
    }

    async function fetchMoreHistory(pageInfo: HistoryResult['pageInfo']) {
        if (!$page.data.resolvedRevision) {
            throw new Error('Unable to resolve repo revision')
        }
        return fetchRepoCommits({
            repoID: $page.data.resolvedRevision.repo.id,
            revision: $page.data.resolvedRevision.commitID,
            filePath: $page.params.path,
            pageInfo,
        })
    }

    let selectedTab: number | null = null
    let historyPanel: HistoryPanel

    // Automatically open history panel when URL contains rev
    $: if ($page.url.searchParams.has('rev')) {
        selectedTab = 0
    }
</script>

<div class:open={selectedTab !== null}>
    <Tabs selected={selectedTab} toggable on:select={selectTab}>
        <TabPanel title="History">
            {#key $page.params.path}
                <HistoryPanel bind:this={historyPanel} {history} fetchMoreHandler={fetchMoreHistory} />
            {/key}
        </TabPanel>
    </Tabs>
</div>

<style lang="scss">
    div {
        position: sticky;
        bottom: 0px;
        background-color: var(--code-bg);
        --align-tabs: flex-start;
        border-top: 1px solid var(--border-color);
        max-height: 50vh;
        overflow: hidden;

        :global(.tabs) {
            height: 100%;
            max-height: 100%;
            overflow: hidden;
        }

        :global(.tabs-header) {
            border-bottom: 1px solid var(--border-color);
        }

        &.open {
            height: 30vh;
        }
    }
</style>
