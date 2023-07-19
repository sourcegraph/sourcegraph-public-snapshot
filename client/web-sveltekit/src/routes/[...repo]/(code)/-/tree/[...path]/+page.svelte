<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'

    import { isErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'

    export let data: PageData

    const { value: treeOrError, set } = createPromiseStore<typeof data.treeEntries.deferred>()
    $: set(data.treeEntries.deferred)
    $: entries = $treeOrError && !isErrorLike($treeOrError) ? $treeOrError.entries : []
</script>

<FileHeader>
    <Icon slot="icon" svgPath={mdiFolderOutline} />
</FileHeader>

<div class="content">
    <h2>Files and directories</h2>
    <ul>
        {#each entries as entry}
            <li>
                <a href={entry.url}>
                    <Icon svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline} inline />
                    {entry.name}
                </a>
            </li>
        {/each}
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
