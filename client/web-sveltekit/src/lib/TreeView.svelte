<svelte:options immutable />

<script lang="ts" generics="N">
    import { createEventDispatcher } from 'svelte'
    import { Key } from 'ts-key-enum'

    import { updateNodeState, type NodeState, type TreeProvider, type TreeState } from './TreeView'
    import TreeViewEntry from './TreeViewEntry.svelte'

    export let treeProvider: TreeProvider<N>
    export let treeState: TreeState
    export let isRoot: boolean

    const dispatch = createEventDispatcher<{ select: HTMLElement }>()

    let element: HTMLElement

    $: entries = treeProvider.getEntries()

    function getFocusedElement(): HTMLElement | null {
        return element.querySelector<HTMLElement>("[role='treeitem'][tabindex='0'][data-node-id]")
    }

    function createNewTreeState(element: HTMLElement, newState: Partial<NodeState>): TreeState {
        return { ...treeState, nodes: updateNodeState(treeState, getNodeID(element), newState)}
    }

    function getNodeID(element: HTMLElement): string {
        return element.dataset.nodeId as string
    }

    function focusNode(element: HTMLElement | null | undefined): void {
        if (element) {
            treeState = { ...treeState, focused: getNodeID(element) }
            element.focus()
        }
    }

    function findSiblingNode(
        element: HTMLElement | null | undefined,
        direction: 'next' | 'previous'
    ): HTMLElement | null {
        if (!element) {
            return null
        }

        // Find next sibling
        let sibling: Element | null = element
        do {
            sibling = direction === 'next' ? element.nextElementSibling : element.previousElementSibling
            if (sibling?.getAttribute('role') === 'treeitem') {
                return sibling as HTMLElement
            }
        } while (sibling)

        return null
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
            next = findSiblingNode(element, 'next')
        }
        if (!next) {
            // Go up
            let nextPossible: HTMLElement|null|undefined = element
            do {
                nextPossible = nextPossible.parentElement?.closest('[role="treeitem"]')
                if (!nextPossible) {
                    break
                }
                next = findSiblingNode(nextPossible, 'next')
            } while (!next)
        }
        return next
    }

    function getPrevFocusableNode(element: HTMLElement | null | undefined): HTMLElement | null {
        if (!element) {
            return null
        }
        // Find previous sibling
        let previous: HTMLElement | null = findSiblingNode(element, 'previous')
        if (previous?.getAttribute('aria-expanded') === 'true') {
            // Find last visible tree item in subtree rooted at the next sibling
            const walker = document.createTreeWalker(previous, NodeFilter.SHOW_ELEMENT, node => (node as HTMLElement).getAttribute("role") === 'treeitem' ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_SKIP)
            let possiblePrev: HTMLElement|null
            while (possiblePrev = walker.lastChild() as HTMLElement|null) {
                previous = possiblePrev
            }
        }
        if (!previous) {
            // Go up
            previous = element.parentElement?.closest('[role="treeitem"]') ?? null
        }
        return previous
    }

    const handledKeys = new Set([Key.ArrowUp, Key.ArrowDown, Key.ArrowLeft, Key.ArrowRight, Key.Enter])

    function handleKeydown(event: KeyboardEvent) {
        if (!handledKeys.has(event.key as Key)) {
            return
        }
        // Prevent arrow keys from scrolling the tree view
        event.preventDefault()
        switch (event.key as Key) {
            case Key.ArrowDown: {
                focusNode(getNextFocusableNode(getFocusedElement()))
                break
            }
            case Key.ArrowUp: {
                focusNode(getPrevFocusableNode(getFocusedElement()))
                break
            }
            case Key.ArrowRight: {
                const focusedElement = getFocusedElement()
                if (focusedElement) {
                    switch (focusedElement.getAttribute('aria-expanded')) {
                        case 'false': {
                            treeState = createNewTreeState(focusedElement, { expanded: true })
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
            case Key.ArrowLeft: {
                const focusedElement = getFocusedElement()
                if (focusedElement) {
                    switch (focusedElement.getAttribute('aria-expanded')) {
                        case 'true': {
                            treeState = createNewTreeState(focusedElement, { expanded: false })
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
            case Key.Enter: {
                const element = getFocusedElement()
                if (element) {
                    dispatch('select', element)
                }
            }
        }
    }

    function handleClick(event: MouseEvent) {
        const element = (event.target as HTMLElement).closest('[role="treeitem"]')
        if (element) {
            dispatch('select', element as HTMLElement)
        }
    }
</script>

<ul
    bind:this={element}
    role={isRoot ? 'tree' : 'group'}
    on:keydown={isRoot ? handleKeydown : undefined}
    on:click={isRoot ? handleClick : undefined}
>
    {#each entries as entry (treeProvider.getKey(entry))}
        <TreeViewEntry {entry} {treeProvider} bind:treeState>
            <svelte:fragment let:entry let:toggle let:expanded>
                <slot {entry} {toggle} {expanded} />
            </svelte:fragment>
        </TreeViewEntry>
    {/each}
</ul>

<style lang="scss">
    ul {
        flex: 1;
        list-style: none;
        padding: 0;
        margin: 0;
        min-height: 0;
    }
</style>
