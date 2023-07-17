<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'

    import { isErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import { asStore } from '$lib/utils'

    import type { PageData } from './$types'

    export let data: PageData

    $: treeDataStatus = asStore(data.treeEntries.deferred)
    $: treeOrError = (!$treeDataStatus.loading && $treeDataStatus.data) || null
    $: entries = treeOrError && !isErrorLike(treeOrError) ? treeOrError.entries : []
</script>

<FileHeader>
    <Icon slot="icon" svgPath={mdiFolderOutline} />
</FileHeader>

<div class="content">
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
    }

    ul {
        list-style: none;
        padding: 0;
        margin: 0;
    }
</style>
