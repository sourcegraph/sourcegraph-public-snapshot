<script lang="ts">

    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'

    import { isErrorLike, type ErrorLike } from '$lib/common'
    import type { TreeFields } from '$lib/graphql/shared'
    import Icon from '$lib/Icon.svelte'

    export let treeOrError: TreeFields | ErrorLike | null

    $: entries = !isErrorLike(treeOrError) && treeOrError ? treeOrError.entries : []
</script>

<table>
    <tr>
        <th>Name</th>
    </tr>
    {#each entries as entry}
        <tr>
            <td>
                <Icon svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline} inline />
                <a href={entry.url}>{entry.name}</a>
            </td>
        </tr>
    {/each}
</table>

<style lang="scss">
    table {
        width: 100%;
        border: 1px solid var(--border-color);
    }

    th {
        padding: 0.25rem;
        border-bottom: 1px solid var(--border-color);
    }

    td {
        background-color: var(--color-bg-1);
        border-bottom: 1px solid var(--border-color);
        padding: 0.25rem;
    }


    a {
        color: var(--body-color);
    }

</style>
