<svelte:options immutable />

<script lang="ts">
    import { mdiFileEyeOutline, mdiMapSearch, mdiWrap, mdiWrapDisabled } from '@mdi/js'
    import { capitalize } from 'lodash'
    import { from } from 'rxjs'
    import { writable } from 'svelte/store'

    import { afterNavigate, goto, preloadData } from '$app/navigation'
    import { page } from '$app/stores'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import { isErrorLike, SourcegraphURL, type LineOrPositionOrRange, pluralize } from '$lib/common'
    import { toGraphQLResult } from '$lib/graphql'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { createBlobDataHandler } from '$lib/repo/blob'
    import FileDiff from '$lib/repo/FileDiff.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import FileIcon from '$lib/repo/FileIcon.svelte'
    import Permalink from '$lib/repo/Permalink.svelte'
    import { createCodeIntelAPI } from '$lib/shared'
    import { formatBytes } from '$lib/utils'
    import { Alert, MenuButton, MenuLink } from '$lib/wildcard'
    import markdownStyles from '$lib/wildcard/Markdown.module.scss'

    import type { PageData, Snapshot } from './$types'
    import FileViewModeSwitcher from './FileViewModeSwitcher.svelte'
    import OpenInCodeHostAction from './OpenInCodeHostAction.svelte'
    import OpenInEditor from '$lib/repo/open-in-editor/OpenInEditor.svelte'
    import { toViewMode, ViewMode } from './util'
    import type { ScrollSnapshot } from '$lib/codemirror/utils'

    export let data: PageData

    export const snapshot: Snapshot<ScrollSnapshot | null> = {
        capture() {
            return cmblob?.getScrollSnapshot() ?? null
        },
        restore(data) {
            initialScrollPosition = data
        },
    }

    const combinedBlobData = createBlobDataHandler()
    const lineWrap = writable<boolean>(false)
    let cmblob: CodeMirrorBlob | null = null
    let blob: Awaited<PageData['blob']> | null = null
    let highlights: Awaited<PageData['highlights']> = ''
    let selectedPosition: LineOrPositionOrRange | null = null
    let initialScrollPosition: ScrollSnapshot | null = null

    $: ({
        repoURL,
        revision,
        resolvedRevision: { commitID },
        repoName,
        filePath,
        settings,
        graphQLClient,
        blameData,
    } = data)
    $: combinedBlobData.set(data.blob, data.highlights)

    $: if (!$combinedBlobData.blobPending) {
        blob = $combinedBlobData.blob
        highlights = $combinedBlobData.highlights
        selectedPosition = SourcegraphURL.from($page.url).lineRange
    }
    $: fileNotFound = $combinedBlobData.blobPending ? null : !$combinedBlobData.blob
    $: fileLoadingError = $combinedBlobData.blobPending ? null : !$combinedBlobData.blob && $combinedBlobData.blobError
    $: isFormatted = !!$combinedBlobData.blob?.richHTML

    $: viewMode = toViewMode($page.url.searchParams.get('view'))
    $: showBlame = viewMode === ViewMode.Blame
    $: showFormatted = isFormatted && viewMode === ViewMode.Default && !showBlame

    $: codeIntelAPI = createCodeIntelAPI({
        settings: setting => (isErrorLike(settings?.final) ? undefined : settings?.final?.[setting]),
        requestGraphQL(options) {
            return from(graphQLClient.query(options.request, options.variables).then(toGraphQLResult))
        },
    })

    afterNavigate(event => {
        // Only restore scroll position when the user used the browser history to navigate back
        // and forth. When the user reloads the page, in which case SvelteKit will also call
        // Snapshot.restore, we don't want to restore the scroll position. Instead we want the
        // selected line (if any) to scroll into view.
        if (event.type !== 'popstate') {
            initialScrollPosition = null
        }
    })

    function viewModeURL(viewMode: ViewMode) {
        switch (viewMode) {
            case ViewMode.Code: {
                const url = SourcegraphURL.from($page.url)
                if (isFormatted) {
                    url.setSearchParameter('view', 'raw')
                } else {
                    url.deleteSearchParameter('view')
                }
                return url.toString()
            }
            case ViewMode.Blame:
                const url = SourcegraphURL.from($page.url)
                url.setSearchParameter('view', 'blame')
                return url.toString()
            case ViewMode.Default:
                return SourcegraphURL.from($page.url).deleteSearchParameter('view').toString()
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
            {#await data.externalServiceType then externalServiceType}
                {#if externalServiceType}
                    <OpenInEditor {externalServiceType} />
                {/if}
            {/await}
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

{#if blob && !blob.binary && !data.compare}
    <div class="file-info">
        <FileViewModeSwitcher
            aria-label="View mode"
            value={viewMode}
            options={isFormatted
                ? [ViewMode.Default, ViewMode.Code, ViewMode.Blame]
                : [ViewMode.Default, ViewMode.Blame]}
            on:preload={event => preloadData(viewModeURL(event.detail))}
            on:change={event => goto(viewModeURL(event.detail), { replaceState: true, keepFocus: true })}
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

<div class="content" class:loading={$combinedBlobData.blobPending} class:compare={!!data.compare} class:fileNotFound>
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
    {:else if $combinedBlobData.blob && showFormatted}
        <div class={`rich ${markdownStyles.markdown}`}>
            {@html $combinedBlobData.blob.richHTML}
        </div>
    {:else if blob}
        <!--
            This ensures that a new CodeMirror instance is created when the file changes.
            This makes the CodeMirror behavior more predictable and avoids issues with
            carrying over state from the previous file.
            Specifically this will make it so that the scroll position is reset to
            `initialScrollPosition` when the file changes.
        -->
        {#key blob.canonicalURL}
            <CodeMirrorBlob
                bind:this={cmblob}
                {initialScrollPosition}
                blobInfo={{
                    ...blob,
                    repoName,
                    commitID,
                    revision: revision ?? '',
                    filePath,
                }}
                {highlights}
                {showBlame}
                blameData={$blameData}
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
        {/key}
    {:else if fileLoadingError}
        <Alert variant="danger">
            Unable to load file data: {fileLoadingError.message}
        </Alert>
    {:else if fileNotFound}
        <div class="circle">
            <Icon svgPath={mdiMapSearch} size={80} />
        </div>
        <h2>File not found</h2>
    {/if}
</div>

<style lang="scss">
    .content {
        display: flex;
        flex-direction: column;
        overflow: auto;
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
