<svelte:options immutable />

<script lang="ts" context="module">
    import { setContext as setContextSvelte, getContext as getContextSvelte } from 'svelte'

    import { updateTreeState, type TreeState, TreeStateUpdate } from './TreeView'

    const CONTEXT_KEY = 'treestore'

    export function setTreeContext(initialState: Writable<TreeState>) {
        return setContextSvelte(CONTEXT_KEY, initialState)
    }

    export function getTreeContext(): Writable<TreeState> {
        return getContextSvelte(CONTEXT_KEY)
    }
</script>

<script lang="ts" generics="N">
    import { createEventDispatcher } from 'svelte'
    import type { Writable } from 'svelte/store'
    import { Key } from 'ts-key-enum'

    import TreeNode from './TreeNode.svelte'
    import type { TreeProvider } from './TreeView'

    export let treeProvider: TreeProvider<N>

    export function scrollSelectedItemIntoView() {
        treeRoot?.querySelector('[aria-selected="true"] [data-treeitem-label]')?.scrollIntoView({ block: 'nearest' })
    }

    const dispatch = createEventDispatcher<{ select: HTMLElement }>()

    let treeState = getTreeContext()
    let treeRoot: HTMLElement

    function getFocusedElement(): HTMLElement | null {
        return treeRoot.querySelector<HTMLElement>("[role='treeitem'][tabindex='0'][data-node-id]")
    }

    function getNodeID(element: HTMLElement): string {
        return element.dataset.nodeId as string
    }

    function focusNode(element: HTMLElement | null | undefined): void {
        if (element) {
            $treeState = updateTreeState($treeState, getNodeID(element), TreeStateUpdate.FOCUS)
            element.focus()
            // We want to scroll actual label into view, not the whole treeitem (including subtree), to
            // prevent the list from jumping.
            element.querySelector('[data-treeitem-label]')?.scrollIntoView({ block: 'nearest' })
        }
    }

    function findSiblingTreeItem(
        element: HTMLElement | null | undefined,
        direction: 'next' | 'previous'
    ): HTMLElement | null {
        if (!element) {
            return null
        }

        // Find next sibling
        let sibling: Element | null = element
        do {
            sibling = direction === 'next' ? sibling.nextElementSibling : sibling.previousElementSibling
            if (sibling?.getAttribute('role') === 'treeitem') {
                return sibling as HTMLElement
            }
        } while (sibling)

        return null
    }

    function findLastDescendantTreeItem(element: HTMLElement | null | undefined): HTMLElement | null {
        if (!element) {
            return null
        }

        const walker = document.createTreeWalker(element, NodeFilter.SHOW_ELEMENT, node => {
            const element = node as HTMLElement
            return element.getAttribute('role') === 'treeitem' ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_SKIP
        })
        let result: HTMLElement | null = null
        let possibleResult: HTMLElement | null = null

        while ((possibleResult = walker.lastChild() as HTMLElement | null)) {
            result = possibleResult
        }

        return result
    }

    function getNextFocusableNode(element: HTMLElement | null | undefined): HTMLElement | null {
        if (!element) {
            return null
        }
        let next: HTMLElement | null = null
        if (element.getAttribute('aria-expanded') === 'true') {
            // Get first child treeitem
            next = element.querySelector('[role="group"] [role="treeitem"]')
        }
        if (!next) {
            // Find next sibling
            next = findSiblingTreeItem(element, 'next')
        }
        if (!next) {
            // Go up till we find an anecstor with a next sibling
            let nextPossible: HTMLElement | null | undefined = element
            do {
                nextPossible = nextPossible.parentElement?.closest('[role="treeitem"]')
                if (!nextPossible) {
                    break
                }
                next = findSiblingTreeItem(nextPossible, 'next')
            } while (!next)
        }
        return next
    }

    function getPrevFocusableNode(element: HTMLElement | null | undefined): HTMLElement | null {
        if (!element) {
            return null
        }
        // Find previous sibling
        let previous: HTMLElement | null = findSiblingTreeItem(element, 'previous')
        if (previous?.getAttribute('aria-expanded') === 'true') {
            previous = findLastDescendantTreeItem(previous)
        }
        if (!previous) {
            // Go up
            previous = element.parentElement?.closest('[role="treeitem"]') ?? null
        }
        return previous
    }

    const handledKeys = new Set([Key.ArrowUp, Key.ArrowDown, Key.ArrowLeft, Key.ArrowRight, Key.Enter])

    // See https://www.w3.org/WAI/ARIA/apg/patterns/treeview/ for more details about event handling
    function handleKeydown(event: KeyboardEvent) {
        if (!handledKeys.has(event.key as Key) || event.ctrlKey || event.metaKey || event.altKey) {
            return
        }
        // Prevent arrow keys from scrolling the tree view
        event.preventDefault()
        switch (event.key as Key) {
            // Focus the next visible tree item
            case Key.ArrowDown: {
                focusNode(getNextFocusableNode(getFocusedElement()))
                break
            }
            // Focus the preceeding visible tree item
            case Key.ArrowUp: {
                focusNode(getPrevFocusableNode(getFocusedElement()))
                break
            }
            // On closed item: expand, don't move focus
            // On open item: move focus to first child menu item
            case Key.ArrowRight: {
                const focusedElement = getFocusedElement()
                if (focusedElement) {
                    switch (focusedElement.getAttribute('aria-expanded')) {
                        case 'false': {
                            $treeState = updateTreeState($treeState, getNodeID(focusedElement), TreeStateUpdate.EXPAND)
                            break
                        }
                        case 'true': {
                            focusNode(focusedElement.querySelector<HTMLElement>('[role="treeitem"][data-node-id]'))
                            break
                        }
                    }
                }
                break
            }
            // On closed item: move focus to parent (if available)
            // On open item: close item, don't move focus
            case Key.ArrowLeft: {
                const focusedElement = getFocusedElement()
                if (focusedElement) {
                    switch (focusedElement.getAttribute('aria-expanded')) {
                        case 'true': {
                            $treeState = updateTreeState(
                                $treeState,
                                getNodeID(focusedElement),
                                TreeStateUpdate.COLLAPSE
                            )
                            break
                        }
                        default: {
                            focusNode(
                                focusedElement.parentElement?.closest<HTMLElement>('[role="treeitem"][data-node-id]')
                            )
                            break
                        }
                    }
                }
                break
            }
            // Select item
            case Key.Enter: {
                const element = getFocusedElement()
                if (element) {
                    dispatch('select', element)
                }
                break
            }
            // Move focus to first visible menu item
            case Key.Home: {
                focusNode(treeRoot.querySelector<HTMLElement>('[role="treeitem"]'))
                break
            }
            // Move focus to first visible menu item
            case Key.End: {
                focusNode(findLastDescendantTreeItem(treeRoot))
                break
            }
        }
    }

    function handleClick(event: MouseEvent) {
        const target = event.target as HTMLElement
        // Only handle clicks on the actual label of the tree item (not on e.g. padding around it)
        if (target.closest('.label')) {
            const item = target.closest<HTMLElement>('[role="treeitem"]')
            if (item) {
                dispatch('select', item)
            }
        }
    }

    $: entries = treeProvider.getEntries()
    // Make first tree item focusable if none is selected/focused
    $: if (!$treeState.focused && entries.length > 0) {
        $treeState = { ...$treeState, focused: treeProvider.getNodeID(entries[0]) }
    }
</script>

<ul bind:this={treeRoot} role="tree" on:keydown={handleKeydown} on:click={handleClick}>
    {#each entries as entry (treeProvider.getNodeID(entry))}
        <TreeNode {entry} {treeProvider}>
            <svelte:fragment let:entry let:toggle let:expanded>
                <slot {entry} {toggle} {expanded} />
            </svelte:fragment>
        </TreeNode>
    {/each}
</ul>

<style lang="scss">
    ul {
        // Padding ensures that focus rings of tree items are not cut off
        padding: 0 0.25rem;

        &,
        :global(ul[role='group']) {
            list-style: none;
            margin: 0;
        }

        :global(ul[role='group']) {
            padding: 0;
        }
    }
</style>
