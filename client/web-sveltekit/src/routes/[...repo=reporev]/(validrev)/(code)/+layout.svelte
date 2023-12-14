<script lang="ts">
    import { onMount } from 'svelte'

    import { afterNavigate, disableScrollHandling } from '$app/navigation'
    import { page } from '$app/stores'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { fetchSidebarFileTree, FileTreeProvider, type FileTreeLoader } from '$lib/repo/api/tree'
    import BottomPanel from '$lib/repo/BottomPanel.svelte'
    import SidebarToggleButton from '$lib/repo/SidebarToggleButton.svelte'
    import { sidebarOpen } from '$lib/repo/stores'
    import Separator, { getSeparatorPosition } from '$lib/Separator.svelte'
    import { scrollAll } from '$lib/stores'

    import type { LayoutData, Snapshot } from './$types'
    import FileTree from './FileTree.svelte'
    import type { Scalars } from '$lib/graphql-operations'

    export let data: LayoutData

    export const snapshot: Snapshot = {
        capture() {
            return {
                bottomPanel: bottomPanel.capture(),
            }
        },
        restore(data) {
            if (data) {
                bottomPanel.restore(data.bottomPanel)
            }
        },
    }

    let bottomPanel: BottomPanel

    const fileTreeLoader: FileTreeLoader = args =>
        fetchSidebarFileTree(args).then(
            ({ root, values }) =>
                new FileTreeProvider({
                    root,
                    values,
                    loader: fileTreeLoader,
                    ...args,
                })
        )
    let treeProvider: FileTreeProvider | null = null

    async function updateFileTreeProvider(repoID: Scalars['ID']['input'], commitID: string, parentPath: string) {
        const result = await data.deferred.fileTree
        if (!result) {
            treeProvider = null
            return
        }
        const { root, values } = result

        // Do nothing if update was called with new arguments in the meantime
        if (
            repoID !== data.resolvedRevision.repo.id ||
            commitID !== data.resolvedRevision.commitID ||
            parentPath !== data.parentPath
        ) {
            return
        }
        treeProvider = new FileTreeProvider({
            root,
            values,
            repoID,
            commitID,
            loader: fileTreeLoader,
        })
    }

    $: ({ revision, parentPath, resolvedRevision } = data)
    $: commitID = resolvedRevision.commitID
    $: repoID = resolvedRevision.repo.id
    // Only update the file tree provider (which causes the tree to rerender) when repo, revision/commit or file path
    // update
    $: updateFileTreeProvider(repoID, commitID, parentPath)

    const sidebarSize = getSeparatorPosition('repo-sidebar', 0.2)
    $: sidebarWidth = `max(200px, ${$sidebarSize * 100}%)`

    onMount(() => {
        // We want the whole page to be scrollable and hide page and repo navigation
        scrollAll.set(true)
        return () => scrollAll.set(false)
    })

    afterNavigate(() => {
        // Prevents SvelteKit from resetting the scroll position to the top
        disableScrollHandling()
    })
</script>

<section>
    <div class="sidebar" class:open={$sidebarOpen} style:min-width={sidebarWidth} style:max-width={sidebarWidth}>
        <h3>
            <SidebarToggleButton />&nbsp; Files
        </h3>
        {#if treeProvider}
            <FileTree revision={revision ?? ''} {treeProvider} selectedPath={$page.params.path ?? ''} />
        {:else}
            <LoadingSpinner center={false} />
        {/if}
    </div>
    {#if $sidebarOpen}
        <Separator currentPosition={sidebarSize} />
    {/if}
    <div class="main">
        <slot />
        <BottomPanel bind:this={bottomPanel} history={data.deferred.codeCommits} />
    </div>
</section>

<style lang="scss">
    section {
        display: flex;
        flex: 1;
        flex-shrink: 0;
        background-color: var(--code-bg);
        min-height: 100vh;
    }

    .sidebar {
        &.open {
            display: flex;
            flex-direction: column;
        }
        display: none;
        overflow: hidden;
        background-color: var(--body-bg);
        padding: 0.5rem;
        padding-bottom: 0;
        position: sticky;
        top: 0px;
        max-height: 100vh;
    }

    .main {
        flex: 1;
        display: flex;
        flex-direction: column;
        min-width: 0;
    }

    h3 {
        display: flex;
        align-items: center;
        margin-bottom: 0.5rem;
    }
</style>
