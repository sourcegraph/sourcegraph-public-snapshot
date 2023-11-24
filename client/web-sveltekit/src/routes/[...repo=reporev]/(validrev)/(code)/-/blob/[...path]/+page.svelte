<svelte:options immutable />

<script lang="ts">
    import { mdiCodeBracesBox, mdiFileCodeOutline } from '@mdi/js'

    import { page } from '$app/stores'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import Permalink from '$lib/repo/Permalink.svelte'

    import FileDiff from '../../../../-/commit/[...revspec]/FileDiff.svelte'

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

    const {
        revision,
        resolvedRevision: { commitID, repo },
        filePath,
        settings,
        graphqlClient,
    } = data
    // We use the latest value here because we want to keep showing the old document while loading
    // the new one.
    const { loading, combinedBlobData, set: setBlobData } = createBlobDataHandler()
    let selectedPosition: LineOrPositionOrRange | null = null

    $: setBlobData(data.deferred.blob, data.deferred.highlights)
    $: blobData = $combinedBlobData.blob
    $: formatted = !!blobData?.richHTML
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

<FileHeader>
    <Icon slot="icon" svgPath={data.deferred.compare ? mdiCodeBracesBox : mdiFileCodeOutline} />
    <svelte:fragment slot="actions">
        {#if data.deferred.compare}
            <span>{data.deferred.compare.revisionToCompare}</span>
        {:else}
            {#if !formatted || showRaw}
                <WrapLinesAction />
            {/if}
            {#if formatted}
                <FormatAction />
            {/if}
            <Permalink resolvedRevision={data.resolvedRevision} />
        {/if}
    </svelte:fragment>
</FileHeader>

<div class="content" class:loading={$loading} class:compare={!!data.deferred.compare}>
    {#if data.deferred.compare}
        {#await data.deferred.compare.diff}
            <LoadingSpinner />
        {:then fileDiff}
            {#if fileDiff}
                <FileDiff {fileDiff} />
            {:else}
                Unable to load iff
            {/if}
        {/await}
    {:else if blobData}
        {#if blobData.richHTML && !showRaw}
            <div class="rich">
                {@html blobData.richHTML}
            </div>
        {:else}
            <CodeMirrorBlob
                blobInfo={{
                    ...blobData,
                    revision: revision ?? '',
                    commitID,
                    repoName: repo.name,
                    filePath,
                }}
                highlights={$combinedBlobData.highlights || ''}
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
