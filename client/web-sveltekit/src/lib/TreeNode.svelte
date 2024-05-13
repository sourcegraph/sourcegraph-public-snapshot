<svelte:options immutable />

<script lang="ts" context="module">
    interface TreeNodeContext {
        level: number
    }
</script>

<script lang="ts" generics="T">
    import { getContext, setContext } from 'svelte'
    import { mdiChevronDown, mdiChevronRight } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import { Button } from '$lib/wildcard'

    import LoadingSpinner from './LoadingSpinner.svelte'
    import { updateTreeState, type TreeProvider, TreeStateUpdate } from './TreeView'
    import { getTreeContext } from './TreeView.svelte'

    export let entry: T
    export let treeProvider: TreeProvider<T>

    let label: HTMLElement | undefined

    $: treeState = getTreeContext()
    $: nodeID = treeProvider.getNodeID(entry)
    $: expandable = treeProvider.isExpandable(entry)
    $: expanded = $treeState.expandedNodes.has(nodeID)
    $: selectable = treeProvider.isSelectable(entry)
    $: selected = $treeState.selected === nodeID
    $: tabindex = $treeState.focused === nodeID ? 0 : -1
    $: children = expandable && expanded ? treeProvider.fetchChildren(entry) : null

    let level = getContext<TreeNodeContext>('tree-node-nesting')?.level ?? 0
    setContext('tree-node-nesting', { level: level + 1 })

    function toggleOpen(expand?: boolean) {
        if (expandable) {
            $treeState = updateTreeState(
                $treeState,
                nodeID,
                expand ?? !expanded ? TreeStateUpdate.EXPANDANDFOCUS : TreeStateUpdate.COLLAPSEANDFOCUS
            )
        }
    }

    $: if (selected && label) {
        const container = label.closest('[role="tree"]')
        if (container) {
            // Only scroll the active tree entry into the 'center' if the selected entry changed
            // by something other than user interaction. If we always 'center' then the tree
            // will "jump" as the user selects an entry with the keyboard or mouse, which is
            // disorienting.
            // But if we never 'center' then going back and forth might position the selected
            // entry at the top or bottom of the scroll container, which is not very visible.
            // So we only 'center' if focus is not on the tree container, which likely means
            // that the user is not interacting with the tree.
            label.scrollIntoView({ block: container.contains(document.activeElement) ? 'nearest' : 'center' })
        }
    }
</script>

<li
    role="treeitem"
    aria-selected={selectable ? selected : undefined}
    aria-expanded={expandable ? expanded : undefined}
    {tabindex}
    data-treeitem
    data-node-id={nodeID}
    style="--tree-node-nested-level: {level}"
>
    <span bind:this={label} class="label" data-treeitem-label>
        <!-- hide the open/close button to preserve alignment with expandable entries -->
        <span class="expandable-icon-container" class:hidden={!expandable}>
            <!-- We have to stop even propagation because the tree root listens for click events for
                 selecting items. We don't want the item to be selected when the open/close button is pressed.
             -->
            <Button
                variant="icon"
                on:click={event => {
                    event.stopPropagation()
                    toggleOpen()
                }}
                tabindex={-1}
            >
                <Icon svgPath={expanded ? mdiChevronDown : mdiChevronRight} inline />
            </Button>
        </span>
        <slot {entry} {expanded} toggle={toggleOpen} />
    </span>
    {#if expanded && children}
        {#await children}
            <div class="loading">
                <LoadingSpinner center={false} />
            </div>
        {:then treeProvider}
            <ul role="group">
                {#each treeProvider.getEntries() as entry (treeProvider.getNodeID(entry))}
                    <svelte:self {entry} {treeProvider} let:entry let:toggle let:expanded>
                        <slot {entry} {toggle} {expanded} />
                    </svelte:self>
                {/each}
            </ul>
        {:catch error}
            <slot name="error" {error} />
        {/await}
    {/if}
</li>

<style lang="scss">
    [role='treeitem'] {
        border-radius: var(--border-radius);

        &[tabindex='0']:focus {
            box-shadow: none;

            > .label {
                box-shadow: var(--focus-box-shadow);
            }
        }
    }

    .loading {
        // Indent with two rem since loading represents next nested level
        margin-left: calc(var(--tree-node-nested-level) * 1.25rem + 1.55rem);
        margin-top: 0.25rem;
    }

    .label {
        display: flex;
        align-items: center;
        padding-right: 0.25rem;
        gap: 0.25rem;
        padding-left: calc(var(--tree-node-nested-level) * 1.25rem + 0.35rem);

        color: var(--tree-node-lable-color, var(--text-body));
        --icon-fill-color: var(--tree-node-expand-icon-color);

        li[data-treeitem][aria-selected='true'] > & {
            color: var(--tree-node-lable-color, var(--body-bg));
        }
    }

    .expandable-icon-container {
        // in order to center/align expandable icon exactly by the item center
        display: flex;
        flex-shrink: 0;
    }

    .hidden {
        visibility: hidden;
    }
</style>
