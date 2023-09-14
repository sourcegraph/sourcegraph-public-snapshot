<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import type { TreeEntryFields } from './api/tree'
    import { replaceRevisionInURL } from '$lib/web'

    export let entries: TreeEntryFields[]
    export let revision: string
</script>

<table>
    <tbody>
        {#each entries as entry}
            <tr>
                <td>
                    <Icon svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline} inline />
                    <a href={replaceRevisionInURL(entry.canonicalURL, revision)}>{entry.name}</a>
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
