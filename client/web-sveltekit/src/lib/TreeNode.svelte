<svelte:options immutable />

<script lang="ts" generics="T">
    import { mdiChevronDown, mdiChevronRight } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import { Button } from '$lib/wildcard'

    import LoadingSpinner from './LoadingSpinner.svelte'
    import { updateNodeState, type TreeProvider } from './TreeView'
    import { getTreeContext } from './TreeView.svelte'

    export let entry: T
    export let treeProvider: TreeProvider<T>

    $: treeState = getTreeContext()
    $: key = treeProvider.getNodeID(entry)
    $: expandable = treeProvider.isExpandable(entry)
    $: expanded = $treeState.nodes[key]?.expanded ?? false
    $: selected = $treeState.nodes[key]?.selected ?? false
    $: tabindex = $treeState.focused === key ? 0 : -1

    $: children = expandable && expanded ? treeProvider.fetchChildren(entry) : null

    function toggleOpen(expand?: boolean) {
        if (expandable) {
            $treeState = {
                focused: key,
                nodes: updateNodeState($treeState, key, { expanded: expand ?? !expanded }),
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
        <!-- hide the open/close button to preserve alignment with expandable entries -->
        <span class:hidden={!expandable}>
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
            <div class="ml-4">
                <LoadingSpinner center={false} />
            </div>
        {:then treeProvider}
            <ul role="group" class="ml-2">
                {#each treeProvider.getEntries() as entry (treeProvider.getNodeID(entry))}
                    <svelte:self {entry} {treeProvider} let:entry let:toggle let:expanded>
                        <slot {entry} {toggle} {expanded} />
                    </svelte:self>
                {/each}
            </ul>
        {/await}
    {/if}
</li>

<style lang="scss">
    li {
        // Margin ensures that focus rings are not covered by preceeding or following elements
        margin: 0.25rem 0;
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

    .hidden {
        visibility: hidden;
    }
</style>
