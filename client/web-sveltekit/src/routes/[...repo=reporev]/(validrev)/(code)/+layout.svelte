<script context="module" lang="ts">
    import { SVELTE_LOGGER, SVELTE_TELEMETRY_EVENTS } from '$lib/telemetry'

    enum TabPanels {
        History,
        References,
    }

    interface Capture {
        selectedTab: number | null
        historyPanel: HistoryCapture
    }

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
    import { mdiChevronDoubleLeft, mdiChevronDoubleRight, mdiHistory, mdiListBoxOutline, mdiHome } from '@mdi/js'
    import { tick } from 'svelte'

    import { afterNavigate, goto } from '$app/navigation'
    import { page } from '$app/stores'
    import { isErrorLike, SourcegraphURL } from '$lib/common'
    import { openFuzzyFinder } from '$lib/fuzzyfinder/FuzzyFinderContainer.svelte'
    import { filesHotkey } from '$lib/fuzzyfinder/keys'
    import Icon from '$lib/Icon.svelte'
    import KeyboardShortcut from '$lib/KeyboardShortcut.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { fetchSidebarFileTree } from '$lib/repo/api/tree'
    import HistoryPanel, { type Capture as HistoryCapture } from '$lib/repo/HistoryPanel.svelte'
    import LastCommit from '$lib/repo/LastCommit.svelte'
    import TabPanel from '$lib/TabPanel.svelte'
    import Tabs from '$lib/Tabs.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Alert, PanelGroup, Panel, PanelResizeHandle, Button } from '$lib/wildcard'
    import type { LastCommitFragment } from '$testing/graphql-type-mocks'

    import type { LayoutData, Snapshot } from './$types'
    import FileTree from './FileTree.svelte'
    import { createFileTreeStore } from './fileTreeStore'
    import type { GitHistory_HistoryConnection, RepoPage_ReferencesLocationConnection } from './layout.gql'
    import ReferencePanel from './ReferencePanel.svelte'
    import RepositoryRevPicker from './RepositoryRevPicker.svelte'

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

    let bottomPanel: Panel
    let fileTreeSidePanel: Panel
    let historyPanel: HistoryPanel
    let selectedTab: number | null = null
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
        // Reset last commit when the query observable changes. Without
        // this we are showing the last commit of the previously selected
        // file/folder until the last commit is loaded.
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

    function toggleFileSidePanel() {
        if (fileTreeSidePanel.isExpanded()) {
            fileTreeSidePanel.collapse()
        } else {
            fileTreeSidePanel.expand()
        }
    }

    function handleGoToRoot(): void {
        // Without this if we go to the root from the before scoped directory
        // (and we were in the root before) fileTreeStore caching omits
        // new file tree provider creating, this would lead to not updating
        // file tree as we go to the root
        fileTreeStore.resetTopPathCache(repoName, resolvedRevision.commitID)
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
    <Panel
        bind:this={fileTreeSidePanel}
        id="sidebar-panel"
        order={1}
        defaultSize={1}
        minSize={15}
        maxSize={35}
        collapsible
        collapsedSize={1}
        let:isCollapsed
    >
        <div class="sidebar" class:collapsed={isCollapsed}>
            <header>
                <div class="sidebar-action-row">
                    <Button variant="secondary" outline size="sm" on:click={toggleFileSidePanel}>
                        <Icon svgPath={!isCollapsed ? mdiChevronDoubleLeft : mdiChevronDoubleRight} inline />
                    </Button>

                    <RepositoryRevPicker
                        repoURL={data.repoURL}
                        revision={data.revision}
                        resolvedRevision={data.resolvedRevision}
                        getRepositoryBranches={data.getRepoBranches}
                        getRepositoryCommits={data.getRepoCommits}
                        getRepositoryTags={data.getRepoTags}
                    />

                    <Tooltip tooltip="Go to the repository root">
                        <Button variant="secondary" outline size="sm">
                            <svelte:fragment slot="custom" let:buttonClass>
                                <a class={buttonClass} href="/{repoName}" on:click={handleGoToRoot}>
                                    <Icon svgPath={mdiHome} inline />
                                </a>
                            </svelte:fragment>
                        </Button>
                    </Tooltip>
                </div>

                <div class="sidebar-action-row">
                    <Button variant="secondary" outline size="sm">
                        <svelte:fragment slot="custom" let:buttonClass>
                            <button
                                class={`${buttonClass} search-files-button`}
                                on:click={() => openFuzzyFinder('files')}
                            >
                                <span>Search files</span>
                                <KeyboardShortcut shorcut={filesHotkey} />
                            </button>
                        </svelte:fragment>
                    </Button>
                </div>
            </header>

            {#if !isCollapsed}
                <div class="sidebar-file-tree">
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
            {/if}
        </div>
    </Panel>

    <PanelResizeHandle id="blob-page-panels-separator" />

    <Panel id="blob-content-panels" order={2}>
        <PanelGroup id="content-panels" direction="vertical">
            <Panel id="main-content-panel" order={1}>
                <slot />
            </Panel>
            <PanelResizeHandle />
            <Panel
                bind:this={bottomPanel}
                id="bottom-actions-panel"
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
        isolation: isolate;
        box-shadow: var(--sidebar-shadow);
        min-width: 2.92rem;
    }

    :global([data-panel-id='blob-content-panels']) {
        z-index: 0;
        isolation: isolate;
    }

    :global([data-panel-resize-handle-id='blob-page-panels-separator']) {
        &::before {
            // Even though side-panel shadow should be rendered over
            // the right content panel, resize handle still should be rendered
            // over any panel elements
            z-index: 2 !important;
        }
    }

    .sidebar {
        height: 100%;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        background-color: var(--color-bg-1);
    }

    .collapsed {
        flex-direction: column;
        align-items: center;

        header {
            flex-wrap: nowrap;
        }

        header,
        .sidebar-action-row {
            flex-direction: column;
            align-items: flex-start;
            gap: 0.5rem;
            width: 100%;
        }

        .search-files-button,
        :global([data-repo-rev-picker-trigger]),
        .sidebar-file-tree {
            display: none;
        }
    }

    header {
        display: flex;
        flex-wrap: wrap;
        gap: 0.25rem;
        align-items: center;
        padding: 0.5rem 0.5rem;

        .sidebar-action-row {
            display: flex;
            flex-basis: 100%;
            align-items: center;
            gap: 0.5rem;
            min-width: 0;
            flex-grow: 1;
        }

        :global([data-repo-rev-picker-trigger]) {
            flex-grow: 1;
        }

        .search-files-button {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 0.25rem;
            min-width: 0;
            flex-grow: 1;

            span {
                white-space: nowrap;
                text-overflow: ellipsis;
                overflow: hidden;
                flex-grow: 1;
                flex-shrink: 1;
                text-align: left;
            }

            kbd {
                display: flex;
                margin: -0.1rem 0;
                letter-spacing: 0.15rem;
            }
        }
    }

    .sidebar-file-tree {
        flex-grow: 1;
        min-height: 0;
        overflow: auto;
        padding: 0.25rem 0 0.5rem 0;
        border-top: 1px solid var(--border-color);
    }

    :global([data-panel-id='main-content-panel']) {
        display: flex;
        flex-direction: column;
        min-width: 0;
        overflow: hidden;
    }

    :global([data-panel-id='bottom-actions-panel']) {
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
