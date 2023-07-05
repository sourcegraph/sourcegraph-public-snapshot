<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import FileHeader from '$lib/repo/ui/FileHeader.svelte'

    import type { PageData } from './$types'
    import FileTable from '$lib/repo/ui/FileTable.svelte'

    export let data: PageData
    $: console.log(data)

    $: treeDataStatus = data.treeEntries
    $: treeOrError = (!$treeDataStatus.loading && $treeDataStatus.data) || null
</script>

<div class="file-header">
<FileHeader commit={data.deferred.history}>
    <Icon slot="icon" svgPath={mdiFolderOutline} />
</FileHeader>
</div>
<div class="content">
    <FileTable {treeOrError}/>

    {#await data.deferred.readmeBlob then blob}
        {#if blob}
            <div class="readme card mt-3">
                <h3 class="header">
                    <Icon svgPath={mdiFileDocumentOutline} inline />
                    {blob.name}
                </h3>
                <div class="readme-content">
                    {#if blob?.richHTML}
                        {@html blob.richHTML}
                    {:else if blob.content}
                        <pre>{blob.content}</pre>
                    {/if}
                </div>
            </div>
        {/if}
    {/await}
</div>

<style lang="scss">
    .file-header {
        padding: 0.5rem;
        background-color: var(--color-bg-1);
        border: 1px solid var(--border-color);
        border-bottom: none;
        padding: 0.5rem;
        border-top-right-radius: var(--border-radius);
        border-top-left-radius: var(--border-radius);
    }

    .content {
        overflow: auto;
    }

    .card {
        .header {
            border: 1px solid var(--border-color);
            background-color: var(--code-bg);
            position: sticky;
            top: 0;
            padding: 0.5rem;
            border-bottom: 1px solid var(--border-color);

        }

        h3.header {
            margin: 0;
        }

        .readme-content {
            border: 1px solid var(--border-color);
            border-top: none;
            background-color: var(--code-bg);
            padding: 0.5rem;

        }
    }
</style>
