<svelte:options immutable />
<script lang="ts">
    import Shimmer from '$lib/Shimmer.svelte'

    import { Button } from "$lib/wildcard"

    import Icon from "$lib/Icon.svelte"
    import { mdiChevronDown, mdiChevronRight } from "@mdi/js"
    import type { TreeProvider} from "$lib/repo/domain/tree"
    import FileTree from "./FileTree.svelte"

    type T = $$Generic
    export let entry: T
    export let treeProvider: TreeProvider<T>

    $: canOpen = treeProvider.canOpen(entry)
    $: url = treeProvider.getURL(entry)
    $: name = treeProvider.getDisplayName(entry)

    let open = treeProvider.canOpen(entry) && treeProvider.isOpen(entry)
    let children: Promise<TreeProvider<T>>|null = null
    const hasLoaded = () => children !== null

    $: children = canOpen && open && !hasLoaded() ? treeProvider.fetchChildren(entry) : null
    $: if (canOpen) console.log(open, hasLoaded(), children, entry)

    function toggleOpen(_open=!open) {
        treeProvider.markOpen(entry, _open)
        open = _open
    }
</script>

<span class="entry">
    <span class:hidden={!canOpen}>
        <Button variant="icon" on:click={() => toggleOpen()}>
            <Icon svgPath={open ? mdiChevronDown : mdiChevronRight} inline />
        </Button>
    </span>
    {#if url}
        <a href={url} on:click={() => toggleOpen(true)}>
            <Icon svgPath={treeProvider.getSVGIconPath(entry, open)} inline />
            {name}
        </a>
    {:else}
        <span>
            <Icon svgPath={treeProvider.getSVGIconPath(entry, open)} inline />&nbsp;
            {name}
        </span>
    {/if}
</span>
{#if open && children}
    <div class="children">
        {#await children}
            <ul>
                <li><Shimmer --height="24px"/></li>
                <li><Shimmer --height="24px"/></li>
            </ul>
        {:then treeProvider}
            <FileTree {treeProvider} activeEntry=""/>
        {/await}
    </div>
{/if}

<style lang="scss">
    .entry {
        display: flex;
        padding: 0.1rem 0;
        align-items: center;
        cursor: pointer;
        border-radius: var(--border-radius);

        &:hover {
            background-color: var(--color-bg-2);
        }

        > * {
            flex-shrink: 0;
        }

        a {
            flex: 1;
            text-overflow: ellipsis;
            overflow: hidden;
            white-space: nowrap;
            text-decoration: none;
        }
    }

    .children {
        ul {
            margin: 0;
            padding: 0;
            list-style: none;
        }
        margin-left: 1rem;
    }

    .hidden {
        visibility: hidden;
    }


    a {
        color: var(--body-color);
        overflow: hidden;
        text-overflow: ellipsis;
    }
</style>
