<svelte:options immutable />

<script lang="ts">
    import { mdiFolderArrowUpOutline, mdiFolderOpenOutline, mdiFolderOutline } from '@mdi/js'
    import { onMount } from 'svelte'

    import { afterNavigate, goto } from '$app/navigation'
    import { getFileIconInfo, DEFAULT_FILE_ICON } from '$lib/fileIcons'
    import Icon from '$lib/Icon.svelte'
    import { type FileTreeProvider, NODE_LIMIT, type FileTreeNodeValue, type TreeEntryFields } from '$lib/repo/api/tree'
    import { getSidebarFileTreeStateForRepo } from '$lib/repo/stores'
    import TreeView, { setTreeContext } from '$lib/TreeView.svelte'
    import { createForwardStore } from '$lib/utils'
    import { replaceRevisionInURL } from '$lib/web'

    export let treeProvider: FileTreeProvider
    export let selectedPath: string
    export let revision: string

    /**
     * Returns the corresponding icon for `entry`
     */
    function getDirectoryIconPath(entry: TreeEntryFields, open: boolean) {
        if (entry === treeRoot) {
            return mdiFolderArrowUpOutline
        }
        return open ? mdiFolderOpenOutline : mdiFolderOutline
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
                if (anchor.dataset.goUp) {
                    let currentTreeProvider = treeProvider
                    treeProvider.fetchParent().then(parentTreeProvider => {
                        if (treeProvider === currentTreeProvider) {
                            treeProvider = parentTreeProvider
                        }
                    })
                } else {
                    goto(anchor.href, { keepFocus: true })
                }
            }
        }
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
     * Takes a file path and makes all intermediate nodes as open, and the last node as selected.
     */
    async function markSelected(path: string) {
        const nodesCopy = new Set($treeState.expandedNodes)

        for (const ancestor of getAncestorPaths(path)) {
            nodesCopy.add(ancestor)
        }
        nodesCopy.add(path)

        $treeState = { focused: path, selected: path, expandedNodes: nodesCopy }
    }

    function scrollSelectedItemIntoView() {
        treeView.scrollSelectedItemIntoView()
    }

    let treeView: TreeView<FileTreeNodeValue>
    let repoID = treeProvider.getRepoID()
    // Since context is only set once when the component is created
    // we need to dynamically sync any changes to the corresponding
    // file tree state store
    const treeState = createForwardStore(getSidebarFileTreeStateForRepo(treeProvider.getRepoID()))
    // Propagating the tree state via context yielded better performance than passing
    // it via props.
    setTreeContext(treeState)

    $: treeRoot = treeProvider.getRoot()
    $: repoID = treeProvider.getRepoID()
    $: treeState.updateStore(getSidebarFileTreeStateForRepo(repoID))
    // Update open and selected nodes when the path changes.
    $: markSelected(selectedPath)

    // Always scroll the selected item into view when we navigate to a different one.
    // NOTE: At the moment this won't always work because the file tree might not be
    // fully loaded after navigation.
    afterNavigate(scrollSelectedItemIntoView)
    // The documentation says afterNavigate will also run on mount but
    // that doesn't seem to be the case
    onMount(scrollSelectedItemIntoView)
</script>

<div tabindex="-1">
    <TreeView bind:this={treeView} {treeProvider} on:select={event => handleSelect(event.detail)}>
        <svelte:fragment let:entry let:expanded>
            {@const isRoot = entry === treeRoot}
            {#if entry === NODE_LIMIT}
                <!-- todo: create alert component -->
                <span class="note">Full list is too long to display. Use search to find a specific file.</span>
            {:else}
                <!--
                    We handle navigation via the TreeView's select event, to preserve the focus state.
                    Using a link here allows us to benefit from data preloading.
                -->
                <a
                    href={replaceRevisionInURL(entry.canonicalURL, revision)}
                    on:click|preventDefault={() => {}}
                    tabindex={-1}
                    data-go-up={isRoot ? true : undefined}
                >
                    {#if entry.isDirectory}
                        <Icon svgPath={getDirectoryIconPath(entry, expanded)} inline />
                    {:else}
                        {@const icon =
                            (entry.__typename === 'GitBlob' && getFileIconInfo(entry.name, entry.languages)?.svg) ||
                            DEFAULT_FILE_ICON}
                        <Icon svgPath={icon.path} inline --color={icon.color} />
                    {/if}
                    {isRoot ? '..' : entry.name}
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
