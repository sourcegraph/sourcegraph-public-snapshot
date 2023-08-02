<script lang="ts">
    import { mdiFolderOutline, mdiFileDocumentOutline } from '@mdi/js'

    import { isErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import { NODE_LIMIT } from '$lib/repo/api/tree'
    import SidebarToggleButton from '$lib/repo/SidebarToggleButton.svelte'
    import { sidebarOpen } from '$lib/repo/stores'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'

    export let data: PageData

    const { value: treeOrError, set: setTree } = createPromiseStore<PageData['deferred']['fileTree']>()
    $: setTree(data.deferred.fileTree)
</script>

{#if !$sidebarOpen}
    <div class="header">
        <SidebarToggleButton />
    </div>
{/if}

<div class="content">
    <h3>Description</h3>
    <p>
        {data.resolvedRevision.repo.description}
    </p>

    {#if $treeOrError && !isErrorLike($treeOrError)}
        <h3>Files and directories</h3>
        <ul class="files">
            {#each $treeOrError.values as entry}
                <li>
                    {#if entry !== NODE_LIMIT}
                        <a
                            data-sveltekit-preload-data={entry.isDirectory ? 'hover' : 'tap'}
                            data-sveltekit-preload-code="hover"
                            href={entry.url}
                            ><Icon svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline} inline />
                            {entry.name}</a
                        >
                    {/if}
                </li>
            {/each}
        </ul>
    {/if}
</div>

<style lang="scss">
    .header {
        padding: 0.5rem;
    }

    .content {
        padding: 1rem;
        overflow: auto;
        flex: 1;
    }

    ul.files {
        padding: 0;
        margin: 0;
        list-style: none;
        columns: 3;
    }
</style>
