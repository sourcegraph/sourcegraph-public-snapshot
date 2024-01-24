<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import Permalink from '$lib/repo/Permalink.svelte'
    import { createPromiseStore } from '$lib/utils'
    import type { TreePage_TreeWithCommitInfo, TreePage_Readme } from './page.gql'

    import type { PageData } from './$types'
    import FileTable from '$lib/repo/FileTable.svelte'
    import Readme from '$lib/repo/Readme.svelte'

    export let data: PageData

    const { value: tree, set: setTree } = createPromiseStore<PageData['treeEntries']>()
    const { value: commitInfo, set: setCommitInfo } = createPromiseStore<Promise<TreePage_TreeWithCommitInfo | null>>()
    const { value: readme, set: setReadme } = createPromiseStore<Promise<TreePage_Readme | null>>()

    $: setTree(data.treeEntries)
    $: setCommitInfo(data.commitInfo)
    $: setReadme(data.readme)
    $: entries = $tree?.entries ?? []
    $: entriesWithCommitInfo = $commitInfo?.entries ?? []
</script>

<svelte:head>
    <title>{data.filePath} - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<FileHeader>
    <Icon slot="icon" svgPath={mdiFolderOutline} />
    <svelte:fragment slot="actions">
        <Permalink commitID={data.resolvedRevision.commitID} />
    </svelte:fragment>
</FileHeader>

<div class="content">
    <FileTable revision={data.revision ?? ''} {entries} commitInfo={entriesWithCommitInfo} />
    {#if $readme}
        <h4 class="header">
            <Icon svgPath={mdiFileDocumentOutline} />
            &nbsp;
            {$readme.name}
        </h4>
        <div class="readme">
            <Readme file={$readme} />
        </div>
    {/if}
</div>

<style lang="scss">
    .content {
        flex: 1;
    }

    .header {
        background-color: var(--body-bg);
        position: sticky;
        top: 2.8rem;
        padding: 0.5rem;
        border-bottom: 1px solid var(--border-color);
        border-top: 1px solid var(--border-color);
        margin: 0;
    }

    .readme {
        padding: 1rem;
        flex: 1;
    }
</style>
