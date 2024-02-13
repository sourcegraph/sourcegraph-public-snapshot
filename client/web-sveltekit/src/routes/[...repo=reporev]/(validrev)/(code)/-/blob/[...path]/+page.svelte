<svelte:options immutable />

<script lang="ts">
    import { mdiCodeBracesBox, mdiFileCodeOutline } from '@mdi/js'

    import { page } from '$app/stores'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import Permalink from '$lib/repo/Permalink.svelte'

    import FileDiff from '$lib/repo/FileDiff.svelte'

    import type { PageData } from './$types'
    import FormatAction from './FormatAction.svelte'
    import WrapLinesAction, { lineWrap } from './WrapLinesAction.svelte'
    import { createCodeIntelAPI, parseQueryAndHash } from '$lib/shared'
    import { goto } from '$app/navigation'
    import { updateSearchParamsWithLineInformation, createBlobDataHandler } from '$lib/repo/blob'
    import { isErrorLike, type LineOrPositionOrRange } from '$lib/common'
    import { from } from 'rxjs'
    import { gql } from '$lib/graphql'

    export let data: PageData

    // We use the latest value here because we want to keep showing the old document while loading
    // the new one.
    const { loading, combinedBlobData, set: setBlobData } = createBlobDataHandler()
    let selectedPosition: LineOrPositionOrRange | null = null

    $: ({
        revision,
        resolvedRevision: { commitID },
        repoName,
        filePath,
        settings,
        graphqlClient,
    } = data)
    $: setBlobData(data.blob, data.highlights)
    $: ({ blob, highlights = '' } = $combinedBlobData)
    $: formatted = !!blob?.richHTML
    $: showRaw = $page.url.searchParams.get('view') === 'raw'
    $: codeIntelAPI = createCodeIntelAPI({
        settings: setting => (isErrorLike(settings.final) ? undefined : settings.final?.[setting]),
        requestGraphQL(options) {
            return from(graphqlClient.query({ query: gql(options.request), variables: options.variables }))
        },
    })
    $: if (!$loading) {
        selectedPosition = parseQueryAndHash($page.url.search, $page.url.hash)
    }
</script>

<svelte:head>
    <title>{filePath} - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<FileHeader>
    <Icon slot="icon" svgPath={data.compare ? mdiCodeBracesBox : mdiFileCodeOutline} />
    <svelte:fragment slot="actions">
        {#if data.compare}
            <span>{data.compare.revisionToCompare}</span>
        {:else}
            {#if !formatted || showRaw}
                <WrapLinesAction />
            {/if}
            {#if formatted}
                <FormatAction />
            {/if}
            <Permalink {commitID} />
        {/if}
    </svelte:fragment>
</FileHeader>

<div class="content" class:loading={$loading} class:compare={!!data.compare}>
    {#if data.compare}
        {#await data.compare.diff}
            <LoadingSpinner />
        {:then fileDiff}
            {#if fileDiff}
                <FileDiff {fileDiff} />
            {:else}
                Unable to load iff
            {/if}
        {/await}
    {:else if blob}
        {#if blob.richHTML && !showRaw}
            <div class="rich">
                {@html blob.richHTML}
            </div>
        {:else}
            <CodeMirrorBlob
                blobInfo={{
                    ...blob,
                    revision: revision ?? '',
                    commitID,
                    repoName: repoName,
                    filePath,
                }}
                {highlights}
                wrapLines={$lineWrap}
                selectedLines={selectedPosition?.line ? selectedPosition : null}
                on:selectline={event => {
                    goto('?' + updateSearchParamsWithLineInformation($page.url.searchParams, event.detail))
                }}
                {codeIntelAPI}
            />
        {/if}
    {/if}
</div>

<style lang="scss">
    .content {
        display: flex;
        overflow-x: auto;
        flex: 1;

        &.compare {
            flex-direction: column;
        }
    }
    .loading {
        filter: blur(1px);
    }

    .rich {
        padding: 1rem;
        overflow: auto;
    }
</style>
