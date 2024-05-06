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

    import { page } from '$app/stores'
    import { afterNavigate, goto } from '$app/navigation'

    import { Alert, PanelGroup, Panel, PanelResizeHandle } from '$lib/wildcard'
    import { isErrorLike, SourcegraphURL } from '$lib/common'
    import { fetchSidebarFileTree } from '$lib/repo/api/tree'
    import { sidebarOpen } from '$lib/repo/stores'
    import type { LastCommitFragment } from '$testing/graphql-type-mocks'

    import Tabs from '$lib/Tabs.svelte'
    import TabPanel from '$lib/TabPanel.svelte'
    import LastCommit from '$lib/repo/LastCommit.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import SidebarToggleButton from '$lib/repo/SidebarToggleButton.svelte'
    import HistoryPanel, { type Capture as HistoryCapture } from '$lib/repo/HistoryPanel.svelte'

    import type { LayoutData, Snapshot } from './$types'
    import FileTree from './FileTree.svelte'
    import { createFileTreeStore } from './fileTreeStore'
    import RepositoryRevPicker from './RepositoryRevPicker.svelte'
    import ReferencePanel from './ReferencePanel.svelte'
    import type { GitHistory_HistoryConnection, RepoPage_ReferencesLocationConnection } from './layout.gql'

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

    let selectedTab: number | null = null
    let bottomPanel: Panel
    let historyPanel: HistoryPanel
    let lastCommit: LastCommitFragment | null
    let commitHistory: GitHistory_HistoryConnection | null
    let references: RepoPage_ReferencesLocationConnection | null
    const fileTreeStore = createFileTreeStore({ fetchFileTreeData: fetchSidebarFileTree })

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
            bottomPanel.collapse()
        }
    })

    async function selectTab(event: { detail: number | null }) {
        trackHistoryPanelTabAction(selectedTab, event.detail)

        if (event.detail === null) {
            const url = new URL($page.url)
            url.searchParams.delete('rev')
            await goto(url, { replaceState: true, keepFocus: true, noScroll: true })
        }
        selectedTab = event.detail
    }

    function handleBottomPanelExpand() {
        if (selectedTab == null) {
            selectedTab = 0
        }
    }

    function handleBottomPanelCollapse() {
        selectedTab = null
    }

    $: {
        if (selectedTab == null) {
            bottomPanel?.collapse()
        } else if (!bottomPanel?.isExpanded()) {
            bottomPanel?.expand()
        }
    }
</script>

<PanelGroup id="blob-page-panels" direction="horizontal">
    {#if $sidebarOpen}
        <Panel id="sidebar-panel" order={1} defaultSize={25} minSize={25} maxSize={35}>
            <div class="sidebar">
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
                        <FileTree
                            {repoName}
                            {revision}
                            treeProvider={$fileTreeStore}
                            selectedPath={$page.params.path ?? ''}
                        />
                    {/if}
                {:else}
                    <LoadingSpinner center={false} />
                {/if}
            </div>
        </Panel>
        <PanelResizeHandle id="blob-page-panels-separator" />
    {/if}

    <Panel id="content-panel" order={2}>
        <PanelGroup id="content-panels" direction="vertical">
            <Panel id="main-content-panel" order={1} defaultSize={90}>
                <slot />
            </Panel>
            <PanelResizeHandle />
            <Panel
                bind:this={bottomPanel}
                id="bottom-tabs-panel"
                order={2}
                defaultSize={1}
                minSize={20}
                maxSize={50}
                collapsible
                collapsedSize={1}
                onExpand={handleBottomPanelExpand}
                onCollapse={handleBottomPanelCollapse}
                let:isCollapsed
            >
                <div class="bottom-panel">
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
                    {#if lastCommit && isCollapsed}
                        <div class="last-commit">
                            <LastCommit {lastCommit} />
                        </div>
                    {/if}
                </div>
            </Panel>
        </PanelGroup>
    </Panel>
</PanelGroup>

<style lang="scss">
    :global([data-panel-group-id='blob-page-panels']) {
        overflow: hidden;
        background-color: var(--code-bg);
    }

    // Forcing the left sidebar panel (file-tree) to be over
    // right content panel to make sure that box-shadow of the
    // sidebar is rendered over content panel.
    :global([data-panel-id='sidebar-panel']) {
        z-index: 1;
        position: relative;
        box-shadow: var(--sidebar-shadow);
    }

    :global([data-panel-id='content-panel']) {
        z-index: 0;
        position: relative;
    }

    .sidebar {
        height: 100%;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        padding: 0.5rem 0.5rem 0 0.5rem;
        background-color: var(--color-bg-1);
    }

    header {
        display: flex;
        gap: 0.5rem;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.5rem;
    }

    h3 {
        display: flex;
        align-items: center;
        margin-bottom: 0;
        flex-shrink: 0;
    }

    :global([data-panel-resize-handle-id='blob-page-panels-separator']) {
        &::before {
            // Even though side-panel shadow should be rendered over
            // the right content panel, resize handle still should be rendered
            // over any panel elements
            z-index: 2 !important;
        }
    }

    // Revision picker trigger button
    header > :global(button) {
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
    }

    :global([data-panel-id='main-content-panel']) {
        display: flex;
        flex-direction: column;
        min-width: 0;
        overflow: hidden;
    }

    :global([data-panel-id='bottom-tabs-panel']) {
        min-height: 2.5625rem; // 41px which is bottom panel compact size
        box-shadow: var(--bottom-panel-shadow);
    }

    .bottom-panel {
        --align-tabs: flex-start;

        display: flex;
        align-items: center;
        flex-flow: row nowrap;
        justify-content: space-between;
        overflow: hidden;
        height: 100%;
        background-color: var(--color-bg-1);
        color: var(--text-body);

        :global([data-tabs]) {
            width: 100%;
        }
    }
</style>
