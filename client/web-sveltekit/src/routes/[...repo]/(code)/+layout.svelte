<script lang="ts">
    import { page } from '$app/stores'
    import FileTree from '$lib/repo/FileTree.svelte'

    import type { PageData } from './$types'
    import Divider, { divider } from '$lib/Divider.svelte'
    import { isErrorLike } from '$lib/common'
    import { FileTreeProvider } from '$lib/repo/api/tree'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import {treePageSidebarOpen} from '$lib/repo/ui/stores'
    import SidebarToggleButton from '$lib/repo/ui/SidebarToggleButton.svelte'
    import Button from '$lib/wildcard/Button.svelte'
    import { mdiFileTreeOutline, mdiHistory } from '@mdi/js'
    import Icon from '$lib/Icon.svelte'
    import { createLocalWritable } from '$lib/stores'
    import HistorySidebar from '$lib/repo/ui/HistorySidebar.svelte'

    export let data: PageData

    function last<T>(arr: T[]): T {
        return arr[arr.length - 1]
    }

    function computeSidebarWidth(ratio: number) {
         return `max(200px, ${ratio * 100}%)`
    }

    let treeProvider: FileTreeProvider|null = null

    // TODO: Keep opened folders open when switching revisions
    async function updateFileTreeProvider(repoName: string, revision: string, fileTreePath: string) {
        treeProvider = null
        const commit = await data.deferred.commitWithTree

        // Do nothing if update was called with new arguments in the meantime
        if (repoName !== data.repoName || revision !== data.revision || fileTreePath !== data.fileTreePath) {
            return
        }

       treeProvider = !isErrorLike(commit) && commit?.tree ? new FileTreeProvider({
            tree: commit.tree,
            repoName,
            revision,
            commitID: commit.oid
        }) : null
    }


    // Only update the tree provider (which causes the file tree to rerender) if the new file tree would be rooted at an
    // ancestor of the current file tree
    $: ({repoName, revision, fileTreePath} = data)
    $: updateFileTreeProvider(repoName, revision, fileTreePath)


    // Controls which sidebar to show
    const selectedSidebar = createLocalWritable("selected-repo-sidebar", "file-tree")
    $: isTreeOrBlobPage = $page.params.path
    $: if (!isTreeOrBlobPage) {
        // Reset sidebar to file tree for repo root pages
        $selectedSidebar = "file-tree"
    }
</script>

<section>
    <div class="sidebar"
        class:open={$treePageSidebarOpen}
        use:divider={{id: "filetree-sidebar", defaultValue: 0.2, compute: computeSidebarWidth}}>
        <div class="actions">
            <SidebarToggleButton />
            <Button variant="secondary" outline={$selectedSidebar === 'commit-history'} size="sm" on:click={() => $selectedSidebar = "file-tree"}>
                <Icon svgPath={mdiFileTreeOutline} inline />
            </Button>
            {#if isTreeOrBlobPage}
                <Button variant="secondary" outline={$selectedSidebar === 'file-tree'} size="sm" on:click={() => $selectedSidebar = "commit-history"}>
                    <Icon svgPath={mdiHistory} inline />
                </Button>
            {/if}
        </div>
        {#if $selectedSidebar === 'file-tree'}
            {#if treeProvider}
                <div class="tree">
                    <FileTree
                        activeEntry={$page.params.path ? last($page.params.path.split('/')) : ''}
                        {treeProvider}
                    />
                </div>
            {:else}
               <LoadingSpinner />
            {/if}
        {/if}
        {#if $selectedSidebar === 'commit-history' && data.resolvedRevision}
           <HistorySidebar repoID={data.resolvedRevision.repo.id} revision={data.revision} path={$page.params.path} />
        {/if}
    </div>
    {#if $treePageSidebarOpen}
         <Divider id="filetree-sidebar" />
    {/if}
    <div class="content">
        <slot />
    </div>
</section>

<style lang="scss">
    section {
        display: flex;
        overflow: hidden;
        flex: 1;
    }

    .sidebar {
        display: none;
        &.open {
            display: flex;
            padding-left: 1rem;
            min-width: 200px;
        }

        overflow: hidden;
        flex-direction: column;

        .actions {
            margin-top: 1rem;
            padding-top: 0.5rem;
            padding-left: 0.5rem;
        }

        .tree {
            margin-top: 1rem;
            overflow: auto;
            padding-right: 0.5rem;
        }
    }

    .content {
        flex: 1;
        overflow: auto;
        display: flex;
        flex-direction: column;
        margin: 1rem;
        margin-bottom: 0;
    }
</style>
