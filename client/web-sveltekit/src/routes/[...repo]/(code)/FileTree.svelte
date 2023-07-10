<svelte:options immutable />

<script lang="ts">
    import { mdiFileCodeOutline, mdiFolderArrowUpOutline, mdiFolderOpenOutline, mdiFolderOutline } from '@mdi/js'

    import type { TreeEntryFields } from '@sourcegraph/shared/src/graphql-operations'

    import { goto } from '$app/navigation'
    import Icon from '$lib/Icon.svelte'
    import { type FileTreeProvider, NODE_LIMIT } from '$lib/repo/api/tree'
    import { getSidebarFileTreeStateForRepo } from '$lib/repo/stores'
    import TreeView, { setTreeContext } from '$lib/TreeView.svelte'
    import { createForwardStore } from '$lib/utils'

    export let treeProvider: FileTreeProvider
    export let selectedPath: string

    /**
     * Returns the corresponding icon for `entry`
     */
    function getIconPath(entry: TreeEntryFields, open: boolean) {
        if (entry === treeRoot) {
            return mdiFolderArrowUpOutline
        }
        if (entry.isDirectory) {
            return open ? mdiFolderOpenOutline : mdiFolderOutline
        }
        return mdiFileCodeOutline
    }

    /**
     * For a given path (e.g. foo/bar/baz) returns a list of ancestor paths (e.g.
     * [foo, foo/bar]
     */
    function getAncestorPaths(path: string) {
        return path
            .split('/')
            .slice(0, -1)
            .map((_, index, segements) => segements.slice(0, index + 1).join('/'))
    }

    /**
     * Navigates to the tree item on selection.
     */
    function handleSelect(element: HTMLElement | null): void {
        if (element) {
            const anchor =
                element.tagName.toLowerCase() === 'a'
                    ? (element as HTMLAnchorElement)
                    : element.querySelector<HTMLAnchorElement>('a')
            if (anchor) {
                goto(anchor.href, { keepFocus: true })
            }
        }
    }

    let repoName = treeProvider.getRepoName()
    // Since context is only set once when the component is created
    // we need to dynamically sync any changes to the corresponding
    // file tree state store
    const treeState = createForwardStore(getSidebarFileTreeStateForRepo(repoName))
    // Propagating the tree state via context yielded better performance than passing
    // it via props.
    setTreeContext(treeState)

    $: repoName = treeProvider.getRepoName()
    $: treeState.updateStore(getSidebarFileTreeStateForRepo(repoName))

    function selectTreeItem(path: string) {
        const nodesCopy = new Set($treeState.expandedNodes)

        for (const ancestor of getAncestorPaths(path)) {
            nodesCopy.add(ancestor)
        }
        nodesCopy.add(path)

        $treeState = { focused: path, selected: path, expandedNodes: nodesCopy }
    }

    // Update open and selected nodes when the path changes.
    $: selectTreeItem(selectedPath)

    $: treeRoot = treeProvider.getRoot()
</script>

<div tabindex="-1">
    <TreeView {treeProvider} on:select={event => handleSelect(event.detail)}>
        <svelte:fragment let:entry let:expanded>
            {#if entry === NODE_LIMIT}
                <!-- todo: create alert component -->
                <span class="note">Full list is too long to display. Use search to find a specific file.</span>
            {:else}
                <!--
                    We handle navigation via the TreeView's select event, to preserve the focus state.
                    Using a link here allows us to benefit from data preloading.
                -->
                <a href={entry.url ?? ''} on:click|preventDefault={() => {}} tabindex={-1}>
                    <Icon svgPath={getIconPath(entry, expanded)} inline />
                    {entry === treeRoot ? '..' : entry.name}
                </a>
            {/if}
        </svelte:fragment>
    </TreeView>
</div>

<style lang="scss">
    div {
        overflow: scroll;

        :global(.treeitem.selectable) > :global(.label) {
            cursor: pointer;
            border-radius: var(--border-radius);

            &:hover {
                background-color: var(--color-bg-2);
            }
        }

        :global(.treeitem.selected) > :global(.label) {
            background-color: var(--color-bg-2);
        }
    }

    a {
        color: var(--body-color);
        flex: 1;
        text-overflow: ellipsis;
        overflow: hidden;
        white-space: nowrap;
        text-decoration: none;
        padding: 0.1rem 0;
    }

    .note {
        color: var(--body-color);
        overflow: hidden;
        border: 1px solid var(--border-color);
        background-color: var(--subtle-bg);
        padding: 0.25rem;
        border-radius: var(--border-radius);
    }
</style>
