<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'

    import type { TreeEntryFields } from '$lib/graphql/shared'
    import Icon from '$lib/Icon.svelte'

    export let entries: TreeEntryFields[]
</script>

<table>
    <tbody>
        {#each entries as entry}
            <tr>
                <td>
                    <Icon svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline} inline />
                    <a href={entry.url}>{entry.name}</a>
                </td>
            </tr>
        {/each}
    </tbody>
</table>

<style lang="scss">
    table {
        width: 100%;
    }

    td {
        background-color: var(--color-bg-1);
        border-bottom: 1px solid var(--border-color);
        padding: 0.25rem 0.5rem;

        &:hover {
            background-color: var(--color-bg-2);
        }
    }

    a {
        color: var(--body-color);

        &:hover {
            color: var(--link-color);
        }
    }
</style>
