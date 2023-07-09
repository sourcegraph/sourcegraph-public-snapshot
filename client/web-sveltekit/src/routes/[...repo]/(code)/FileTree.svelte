<script lang="ts">
    import { mdiFileCodeOutline, mdiFolderArrowUpOutline, mdiFolderOpenOutline, mdiFolderOutline } from '@mdi/js'
    import { writable } from 'svelte/store'

    import type { TreeEntryFields } from '@sourcegraph/shared/src/graphql-operations'

    import { goto } from '$app/navigation'
    import Icon from '$lib/Icon.svelte'
    import type { FileTreeProvider } from '$lib/repo/api/tree'
    import type { TreeState } from '$lib/TreeView'
    import TreeView, { setTreeContext } from '$lib/TreeView.svelte'

    export let treeProvider: FileTreeProvider
    export let selectedPath: string

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

    function gotoEntry(anchor: HTMLAnchorElement): void {
        goto(anchor.href, { keepFocus: true })
    }

    function handleInteraction(element: EventTarget | null): void {
        if (element) {
            const target = element as HTMLElement
            const anchor =
                target.tagName.toLowerCase() === 'a'
                    ? (target as HTMLAnchorElement)
                    : target.querySelector<HTMLAnchorElement>('a')
            if (anchor) {
                gotoEntry(anchor)
            }
        }
    }

    $: treeRoot = treeProvider.getRoot()

    let treeState: TreeState = {
        focused: '',
        nodes: {},
    }
    let treeStateStore = writable(treeState)
    setTreeContext(treeStateStore)
    let currentlySelectedPath: string

    // Update open and selected nodes when the path changes
    $: if (currentlySelectedPath !== selectedPath) {
        const nodesCopy = { ...$treeStateStore.nodes }
        for (const ancestor of getAncestorPaths(selectedPath)) {
            nodesCopy[ancestor] = { ...nodesCopy[ancestor], expanded: true }
        }
        nodesCopy[selectedPath] = { ...nodesCopy[selectedPath], selected: true, expanded: true }
        if (nodesCopy[currentlySelectedPath]) {
            nodesCopy[currentlySelectedPath] = { ...nodesCopy[currentlySelectedPath], selected: false }
        }
        currentlySelectedPath = selectedPath
        $treeStateStore = { focused: selectedPath, nodes: nodesCopy }
    }
</script>

<div tabindex="-1">
    <TreeView {treeProvider} isRoot on:select={event => handleInteraction(event.detail)}>
        <svelte:fragment let:entry let:expanded>
            <!-- we progamatically handle navigation to preserve the focus state (see gotoEntry) -->
            <a href={entry.url ?? ''} on:click|preventDefault={event => handleInteraction(event.target)} tabindex={-1}>
                <Icon svgPath={getIconPath(entry, expanded)} inline />
                {entry === treeRoot ? '..' : entry.name}
            </a>
        </svelte:fragment>
    </TreeView>
</div>

<style lang="scss">
    div {
        overflow: scroll;

        :global(.label) {
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
</style>
