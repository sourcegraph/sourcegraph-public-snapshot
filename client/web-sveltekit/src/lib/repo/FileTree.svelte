<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'
    import VirtualList from '@sveltejs/svelte-virtual-list'

    import { isErrorLike, type ErrorLike } from '$lib/common'
    import type { TreeFields } from '$lib/graphql/shared'

    import Icon from '$lib/Icon.svelte'

    export let treeOrError: TreeFields | ErrorLike | null
    export let activeEntry: string
    export let commitData: string | null = null

    function scrollIntoView(node: HTMLElement, scroll: boolean) {
        if (scroll) {
            console.log(scroll, node)
            node.scrollIntoView()
        }
    }

    $: entries = !isErrorLike(treeOrError) && treeOrError ? treeOrError.entries : []
</script>

<slot name="title">
    <h3>Files</h3>
</slot>
<ul>
    <VirtualList items={entries} let:item={entry}>
        <li class:active={entry.name === activeEntry} use:scrollIntoView={entry.name === 'activeEntry'}>
            <a href={entry.url}>
                <span>
                    <Icon svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline} inline />
                </span>
                <span class="name">{entry.name}</span>
            </a>
            {#if commitData}
                <span class="ml-5">{commitData}</span>
            {/if}
        </li>
    </VirtualList>
</ul>

<style lang="scss">
    ul {
        flex: 1;
        list-style: none;
        padding: 0;
        margin: 0;
        overflow: auto;
        min-height: 0;
    }

    li {
        display: flex;

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
