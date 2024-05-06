<script context="module" lang="ts">
    import { SVELTE_LOGGER, SVELTE_TELEMETRY_EVENTS } from '$lib/telemetry'

    // Not ideal solution, [TODO] Improve Tabs component API in order
    // to expose more info about nature of switch tab / close tab actions
    function trackHistoryPanelTabAction(selectedTab: number | null, nextSelectedTab: number | null) {
        if (nextSelectedTab === 0) {
            SVELTE_LOGGER.log(SVELTE_TELEMETRY_EVENTS.ShowHistoryPanel)
            return
        }

        if (nextSelectedTab === null && selectedTab == 0) {
            SVELTE_LOGGER.log(SVELTE_TELEMETRY_EVENTS.HideHistoryPanel)
            return
        }
    }
</script>

<script lang="ts">
    import { mdiHistory, mdiListBoxOutline } from '@mdi/js'
    import { tick } from 'svelte'

    import { afterNavigate, goto } from '$app/navigation'
    import { page } from '$app/stores'
    import { isErrorLike, SourcegraphURL } from '$lib/common'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { fetchSidebarFileTree } from '$lib/repo/api/tree'
    import HistoryPanel, { type Capture as HistoryCapture } from '$lib/repo/HistoryPanel.svelte'
    import LastCommit from '$lib/repo/LastCommit.svelte'
    import SidebarToggleButton from '$lib/repo/SidebarToggleButton.svelte'
    import { sidebarOpen } from '$lib/repo/stores'
    import Separator, { getSeparatorPosition } from '$lib/Separator.svelte'
    import TabPanel from '$lib/TabPanel.svelte'
    import Tabs from '$lib/Tabs.svelte'
    import { Alert } from '$lib/wildcard'
    import type { LastCommitFragment } from '$testing/graphql-type-mocks'

    import type { LayoutData, Snapshot } from './$types'
    import FileTree from './FileTree.svelte'
    import { createFileTreeStore } from './fileTreeStore'

    import type { GitHistory_HistoryConnection, RepoPage_ReferencesLocationConnection } from './layout.gql'
    import RepositoryRevPicker from './RepositoryRevPicker.svelte'
    import ReferencePanel from './ReferencePanel.svelte'

    enum TabPanels {
        History,
        References,
    }

    interface Capture {
        selectedTab: number | null
        historyPanel: HistoryCapture
    }

    export let data: LayoutData

    export const snapshot: Snapshot<Capture> = {
        capture() {
            return {
                selectedTab,
                historyPanel: historyPanel?.capture(),
            }
        },
        async restore(data) {
            selectedTab = data.selectedTab
            // Wait until DOM was updated to possibly show the history panel
            await tick()

            // Restore history panel state if it is open
            if (data.historyPanel) {
                historyPanel?.restore(data.historyPanel)
            }
        },
    }

    async function selectTab(event: { detail: number | null }) {
        trackHistoryPanelTabAction(selectedTab, event.detail)

        if (event.detail === null) {
            const url = new URL($page.url)
            url.searchParams.delete('rev')
            await goto(url, { replaceState: true, keepFocus: true, noScroll: true })
        }
        selectedTab = event.detail
    }

    const fileTreeStore = createFileTreeStore({ fetchFileTreeData: fetchSidebarFileTree })
    let selectedTab: number | null = null
    let historyPanel: HistoryPanel
    let commitHistory: GitHistory_HistoryConnection | null
    let references: RepoPage_ReferencesLocationConnection | null
    let lastCommit: LastCommitFragment | null

    $: ({ revision = '', parentPath, repoName, resolvedRevision } = data)
    $: fileTreeStore.set({ repoName, revision: resolvedRevision.commitID, path: parentPath })
    $: commitHistoryQuery = data.commitHistory
    $: lastCommitQuery = data.lastCommit
    $: if (!!commitHistoryQuery) {
        // Reset commit history when the query observable changes. Without
        // this we are showing the commit history of the previously selected
        // file/folder until the new commit history is loaded.
        commitHistory = null
    }

    $: if (!!lastCommitQuery) {
        // Reset commit history when the query observable changes. Without
        // this we are showing the commit history of the previously selected
        // file/folder until the new commit history is loaded.
        lastCommit = null
    }

    $: commitHistory = $commitHistoryQuery?.data?.repository?.commit?.ancestors ?? null
    $: lastCommit = $lastCommitQuery?.data?.repository?.lastCommit?.ancestors?.nodes[0] ?? null

    const sidebarSize = getSeparatorPosition('repo-sidebar', 0.2)
    $: sidebarWidth = `max(320px, ${$sidebarSize * 100}%)`

    // The observable query to fetch references (due to infinite scrolling)
    $: sgURL = SourcegraphURL.from($page.url)
    $: selectedLine = sgURL.lineRange
    $: referenceQuery =
        sgURL.viewState === 'references' && selectedLine?.line ? data.getReferenceStore(selectedLine) : null
    $: references = $referenceQuery?.data?.repository?.commit?.blob?.lsif?.references ?? null
    $: referencesLoading = ((referenceQuery && !references) || $referenceQuery?.fetching) ?? false

    afterNavigate(async () => {
        // We need to wait for referenceQuery to be updated before checking its state
        await tick()

        // todo(fkling): Figure out a proper way to represent bottom panel state
        if (sgURL.viewState === 'references') {
            selectedTab = TabPanels.References
        } else if ($page.url.searchParams.has('rev')) {
            // The file view/history panel use the 'rev' parameter to specify the commit to load
            selectedTab = TabPanels.History
        } else if (selectedTab === TabPanels.References) {
            // Close references panel when navigating to a URL that doesn't have the 'references' view state
            selectedTab = null
        }
    })
</script>

<section>
    <div class="sidebar" class:open={$sidebarOpen} style:min-width={sidebarWidth} style:max-width={sidebarWidth}>
        <header>
            <h3>
                <SidebarToggleButton />&nbsp; Files
            </h3>
            <RepositoryRevPicker
                repoURL={data.repoURL}
                revision={data.revision}
                resolvedRevision={data.resolvedRevision}
                getRepositoryBranches={data.getRepoBranches}
                getRepositoryCommits={data.getRepoCommits}
                getRepositoryTags={data.getRepoTags}
            />
        </header>
        {#if $fileTreeStore}
            {#if isErrorLike($fileTreeStore)}
                <Alert variant="danger">
                    Unable to fetch file tree data:
                    {$fileTreeStore.message}
                </Alert>
            {:else}
                <FileTree {repoName} {revision} treeProvider={$fileTreeStore} selectedPath={$page.params.path ?? ''} />
            {/if}
        {:else}
            <LoadingSpinner center={false} />
        {/if}
    </div>
    {#if $sidebarOpen}
        <Separator currentPosition={sidebarSize} />
    {/if}
    <div class="main">
        <slot />
        <div class="bottom-panel" class:open={selectedTab !== null}>
            <Tabs selected={selectedTab} toggable on:select={selectTab}>
                <TabPanel title="History" icon={mdiHistory}>
                    {#key $page.params.path}
                        <HistoryPanel
                            bind:this={historyPanel}
                            history={commitHistory}
                            loading={$commitHistoryQuery?.fetching ?? true}
                            fetchMore={commitHistoryQuery.fetchMore}
                            enableInlineDiff={$page.data.enableInlineDiff}
                            enableViewAtCommit={$page.data.enableViewAtCommit}
                        />
                    {/key}
                </TabPanel>
                <TabPanel title="References" icon={mdiListBoxOutline}>
                    <ReferencePanel
                        connection={references}
                        loading={referencesLoading}
                        on:more={referenceQuery?.fetchMore}
                    />
                </TabPanel>
            </Tabs>
            {#if lastCommit}
                <div class="last-commit">
                    <LastCommit {lastCommit} />
                </div>
            {/if}
        </div>
    </div>
</section>

<style lang="scss">
    section {
        display: flex;
        flex: 1;
        overflow: hidden;
        background-color: var(--code-bg);
    }

    header {
        display: flex;
        gap: 0.5rem;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.5rem;
    }

    .sidebar {
        &.open {
            display: flex;
            flex-direction: column;
        }
        display: none;
        overflow: hidden;
        background-color: var(--color-bg-1);
        padding: 0.5rem;
        padding-bottom: 0;
        box-shadow: var(--sidebar-shadow);
        z-index: 2;
    }

    .main {
        flex: 1;
        display: flex;
        flex-direction: column;
        min-width: 0;
        overflow: hidden;
    }

    h3 {
        display: flex;
        align-items: center;
        margin-bottom: 0;
        flex-shrink: 0;
    }

    // Revision picker trigger button
    header > :global(button) {
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
    }

    .bottom-panel {
        --align-tabs: flex-start;

        display: flex;
        align-items: center;
        flex-flow: row nowrap;
        justify-content: space-between;
        overflow: hidden;
        border-top: 1px solid var(--border-color);
        box-shadow: var(--bottom-panel-shadow);
        background-color: var(--color-bg-1);
        color: var(--text-body);
        // Applying z-index to the bottom panel allows its shadow to cascade correctly on the code but still say behind the left panel
        z-index: 1;

        &.open {
            height: 32vh;
            // Disable flex layout so that tabs simply fill the available space
            display: block;

            .last-commit {
                display: none;
            }
        }
    }
</style>
