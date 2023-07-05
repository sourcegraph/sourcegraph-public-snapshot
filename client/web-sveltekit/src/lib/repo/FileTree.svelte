<svelte:options immutable/>
<script lang="ts">
    import type { TreeProvider } from './domain/tree'

    import FileTreeEntry from './FileTreeEntry.svelte'

    type T = $$Generic<TreeEntry>
    export let treeProvider: TreeProvider<T>
    export let activeEntry: string

    function scrollIntoView(node: HTMLElement, scroll: boolean) {
        if (scroll) {
            console.log(scroll, node)
            node.scrollIntoView()
        }
    }

    $: entries = treeProvider.getEntries()
</script>

<slot name="title"></slot>
<ul>
    {#each entries as entry (treeProvider.getKey(entry))}
        <li class:active={entry.id === activeEntry} use:scrollIntoView={entry.id === activeEntry}>
            <FileTreeEntry {entry} {treeProvider} />
        </li>
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

    li {
        a {
            flex: 1;
            white-space: nowrap;
            color: var(--body-color);
            text-decoration: none;
            padding: 0.25rem;
        }

        &:hover {
            a {
                background-color: var(--color-bg-2);
            }

            .name {
                text-decoration: underline;
            }
        }

        &.active a {
            background-color: var(--color-bg-3);
        }
    }

    span {
        position: sticky;
        left: 0;
        background-color: inherit;
    }
</style>
