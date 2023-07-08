<svelte:options immutable />

<script lang="ts" generics="T">
    import { mdiChevronDown, mdiChevronRight } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import { Button } from '$lib/wildcard'

    import LoadingSpinner from './LoadingSpinner.svelte'
    import { updateNodeState, type NodeState, type TreeProvider, type TreeState } from './TreeView'
    import TreeView from './TreeView.svelte'

    export let entry: T
    export let treeState: TreeState
    export let treeProvider: TreeProvider<T>

    $: key = treeProvider.getKey(entry)
    $: expandable = treeProvider.isExpandable(entry)
    $: ({ expanded, selected } = treeState.nodes[key] ?? { expanded: false, selected: false })
    $: tabindex = treeState.focused === key ? 0 : -1

    $: children = expandable && expanded ? treeProvider.fetchChildren(entry) : null

    function toggleOpen(expand?: boolean) {
        if (expandable) {
            treeState = {
                focused: key,
                nodes: updateNodeState(treeState, key, { expanded: expand ?? !expanded }) ,
            }
        }
    }

    function scrollIntoView(node: HTMLElement, scroll: boolean) {
        if (scroll) {
            node.scrollIntoView()
        }
    }
</script>

<li
    class="treeitem"
    class:selected
    use:scrollIntoView={selected}
    role="treeitem"
    aria-selected={selected}
    aria-expanded={expandable ? expanded : undefined}
    {tabindex}
    data-node-id={key}
>
    <span class="label">
        <span class:hidden={!expandable}>
            <Button variant="icon" on:click={event => {event.stopPropagation(); toggleOpen()}} tabindex={-1}>
                <Icon svgPath={expanded ? mdiChevronDown : mdiChevronRight} inline />
            </Button>
        </span>
        <slot {entry} {expanded} toggle={toggleOpen} />
    </span>
    {#if expanded && children}
        <div class="children">
            {#await children}
                <LoadingSpinner />
            {:then treeProvider}
                <TreeView {treeProvider} bind:treeState isRoot={false} let:entry let:toggle let:expanded>
                    <slot {entry} {toggle} {expanded} />
                </TreeView>
            {/await}
        </div>
    {/if}
</li>

<style lang="scss">
    li {
        margin: 0 0.2rem;
        border-radius: var(--border-radius);

        &[aria-expanded='true']:focus {
            box-shadow: none;

            > .label {
                box-shadow: var(--focus-box-shadow);
            }
        }
    }

    .label {
        display: flex;
        align-items: center;
    }

    .children {
        margin-left: 1rem;
    }

    .hidden {
        visibility: hidden;
    }
</style>
