<svelte:options immutable />

<script lang="ts">
    import { mdiFileEyeOutline, mdiMapSearch, mdiWrap, mdiWrapDisabled } from '@mdi/js'
    import { from } from 'rxjs'
    import { toViewMode, ViewMode } from './util'

    import { goto } from '$app/navigation'
    import { page } from '$app/stores'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import { isErrorLike, SourcegraphURL, type LineOrPositionOrRange, pluralize } from '$lib/common'
    import { toGraphQLResult } from '$lib/graphql'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { createBlobDataHandler } from '$lib/repo/blob'
    import FileDiff from '$lib/repo/FileDiff.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import Permalink from '$lib/repo/Permalink.svelte'
    import { createCodeIntelAPI } from '$lib/shared'
    import { Alert, MenuButton, MenuLink } from '$lib/wildcard'
    import markdownStyles from '$lib/wildcard/Markdown.module.scss'

    import type { PageData } from './$types'
    import FileIcon from '$lib/repo/FileIcon.svelte'
    import { formatBytes } from '$lib/utils'
    import FileViewModeSwitcher from './FileViewModeSwitcher.svelte'
    import { capitalize } from 'lodash'
    import OpenInCodeHostAction from './OpenInCodeHostAction.svelte'
    import { writable } from 'svelte/store'

    export let data: PageData

    const combinedBlobData = createBlobDataHandler()
    let selectedPosition: LineOrPositionOrRange | null = null
    const lineWrap = writable<boolean>(false)

    $: ({
        repoURL,
        revision,
        resolvedRevision: { commitID },
        repoName,
        filePath,
        settings,
        graphQLClient,
    } = data)
    $: viewMode = toViewMode($page.url.searchParams.get('view'))
    $: combinedBlobData.set(data.blob, data.highlights)
    $: ({ blob, highlights, blobPending } = $combinedBlobData)
    $: isFormatted = !!blob?.richHTML
    $: fileNotFound = !blob && !blobPending
    $: fileLoadingError = (!blobPending && !blob && $combinedBlobData.blobError) || null
    $: showRaw = $page.url.searchParams.get('view') === 'raw'
    $: codeIntelAPI = createCodeIntelAPI({
        settings: setting => (isErrorLike(settings?.final) ? undefined : settings?.final?.[setting]),
        requestGraphQL(options) {
            return from(graphQLClient.query(options.request, options.variables).then(toGraphQLResult))
        },
    })
    $: if (!blobPending) {
        // Update selected position as soon as blob is loaded
        selectedPosition = SourcegraphURL.from($page.url).lineRange
    }

    function changeViewMode({ detail: viewMode }: { detail: ViewMode }) {
        switch (viewMode) {
            case ViewMode.Code: {
                const url = SourcegraphURL.from($page.url)
                if (isFormatted) {
                    url.setSearchParameter('view', 'raw')
                } else {
                    url.deleteSearchParameter('view')
                }
                goto(url.toString(), { replaceState: true, keepFocus: true })
                break
            }
            case ViewMode.Blame:
                // TODO: Implement
                break
            case ViewMode.Default:
                goto(SourcegraphURL.from($page.url).deleteSearchParameter('view').toString(), {
                    replaceState: true,
                    keepFocus: true,
                })
                break
        }
    }
</script>

<svelte:head>
    <title>{filePath} - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<!-- Note: Splitting this at this level is not great but Svelte doesn't allow to conditionally render slots (yet) -->
{#if data.compare}
    <FileHeader>
        <FileIcon slot="icon" file={blob} inline />
        <svelte:fragment slot="actions">
            <span>{data.compare.revisionToCompare}</span>
        </svelte:fragment>
    </FileHeader>
{:else}
    <FileHeader>
        <FileIcon slot="icon" file={blob} inline />
        <svelte:fragment slot="actions">
            {#if blob}
                <OpenInCodeHostAction data={blob} />
            {/if}
            <Permalink {commitID} />
        </svelte:fragment>
        <svelte:fragment slot="actionmenu">
            <MenuLink href="{repoURL}/-/raw/{filePath}" target="_blank">
                <Icon svgPath={mdiFileEyeOutline} inline /> View raw
            </MenuLink>
            <MenuButton
                on:click={() => lineWrap.update(wrap => !wrap)}
                disabled={viewMode === ViewMode.Default && isFormatted}
            >
                <Icon svgPath={$lineWrap ? mdiWrap : mdiWrapDisabled} inline />
                {$lineWrap ? 'Disable' : 'Enable'} wrapping long lines
            </MenuButton>
        </svelte:fragment>
    </FileHeader>
{/if}

{#if !blobPending && blob && !blob.binary && !data.compare}
    <div class="file-info">
        <FileViewModeSwitcher
            aria-label="View mode"
            value={viewMode}
            options={isFormatted ? [ViewMode.Default, ViewMode.Code] : [ViewMode.Default]}
            on:change={changeViewMode}
        >
            <svelte:fragment slot="label" let:value>
                {value === ViewMode.Default ? (isFormatted ? 'Formatted' : 'Code') : capitalize(value)}
            </svelte:fragment>
        </FileViewModeSwitcher>
        <code>
            {blob.totalLines}
            {pluralize('line', blob.totalLines)} · {formatBytes(blob.byteSize)}
        </code>
    </div>
{/if}

<div class="content" class:loading={blobPending} class:compare={!!data.compare} class:fileNotFound>
    {#if !$combinedBlobData.highlightsPending && $combinedBlobData.highlightsError}
        <Alert variant="danger">
            Unable to load syntax highlighting: {$combinedBlobData.highlightsError.message}
        </Alert>
    {/if}
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
            <div class={`rich ${markdownStyles.markdown}`}>
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
                on:selectline={({ detail: range }) => {
                    goto(
                        SourcegraphURL.from($page.url.searchParams)
                            .setLineRange(range ? { line: range.line, endLine: range.endLine } : null)
                            .deleteSearchParameter('popover').search
                    )
                }}
                {codeIntelAPI}
            />
        {/if}
    {:else if !blobPending}
        {#if fileLoadingError}
            <Alert variant="danger">
                Unable to load file data: {fileLoadingError.message}
            </Alert>
        {:else if fileNotFound}
            <div class="circle">
                <Icon svgPath={mdiMapSearch} size={80} />
            </div>
            <h2>File not found</h2>
        {/if}
    {/if}
</div>

<style lang="scss">
    .content {
        display: flex;
        flex-direction: column;
        overflow-x: auto;
        flex: 1;

        &.compare {
            flex-direction: column;
        }

        &.fileNotFound {
            background-color: var(--body-bg);
            flex-direction: column;
            align-items: center;
        }
    }

    .file-info {
        padding: 0.5rem 0.5rem;
        color: var(--text-muted);
        display: flex;
        gap: 1rem;
        align-items: baseline;
    }

    .loading {
        filter: blur(1px);
    }

    .rich {
        padding: 1rem;
        overflow: auto;
        max-width: 50rem;
    }

    .circle {
        background-color: var(--color-bg-2);
        border-radius: 50%;
        padding: 1.5rem;
        margin: 1rem;
    }

    .actions {
        margin-left: auto;
    }
</style>
