<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolder, mdiFolderOutline } from '@mdi/js'

    import { page } from '$app/stores'
    import { isErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'

    import type { PageData } from './$types'

    export let data: PageData

    $: treeDataStatus = data.treeEntries
    $: treeOrError = (!$treeDataStatus.loading && $treeDataStatus.data) || null
    $: entries = treeOrError && !isErrorLike(treeOrError) ? treeOrError.entries : []
</script>

<div class="content">
    <h1>
        <Icon svgPath={mdiFolder} />
        <span class="ml-2">{$page.params.path}</span>
    </h1>
    <h2>Files and directories</h2>
    <ul>
        {#if treeOrError}
            {#each entries as entry}
                <li>
                    <a href={entry.url}>
                        <Icon svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline} inline />
                        {entry.name}
                    </a>
                </li>
            {/each}
        {/if}
    </ul>
</div>

<style lang="scss">
    .content {
        padding: 1rem;
        overflow: auto;
    }

    h1 {
        display: flex;
        align-items: center;
    }

    ul {
        list-style: none;
        padding: 0;
        margin: 0;
    }
</style>
