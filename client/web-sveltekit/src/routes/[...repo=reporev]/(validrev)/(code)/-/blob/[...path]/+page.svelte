<svelte:options immutable />

<script lang="ts">
    // @sg EnableRollout

    import { dirname } from 'path'

    import { onMount } from 'svelte'

    import { afterNavigate, beforeNavigate } from '$app/navigation'
    import { SVELTE_LOGGER, SVELTE_TELEMETRY_EVENTS } from '$lib/telemetry'
    import { TELEMETRY_V2_RECORDER } from '$lib/telemetry2'

    import { getRepositoryPageContext } from '../../../../../context'

    import type { PageData, Snapshot } from './$types'
    import DiffView from './DiffView.svelte'
    import FileView from './FileView.svelte'

    export let data: PageData

    export const snapshot: Snapshot = {
        capture() {
            return {
                fileView: fileView?.capture(),
            }
        },
        restore(data) {
            if (data.fileView) {
                fileView?.restore(data.fileView)
            }
        },
    }

    async function updateRepositoryContextFromBlob(blob: PageData['blob']) {
        try {
            const fileData = await blob
            // Data is not stale
            if (blob === data.blob) {
                const fileLanguage = fileData?.languages?.[0]
                if (fileLanguage) {
                    repositoryContext.update(context => ({ ...context, fileLanguage }))
                }
            }
        } catch (error) {
            // Do nothing
        }
    }

    const repositoryContext = getRepositoryPageContext()
    let fileView: FileView

    afterNavigate(() => {
        repositoryContext.set({
            directoryPath: dirname(data.filePath),
            filePath: data.filePath,
        })
        updateRepositoryContextFromBlob(data.blob)
    })
    beforeNavigate(() => {
        repositoryContext.set({})
    })

    onMount(() => {
        SVELTE_LOGGER.logViewEvent(SVELTE_TELEMETRY_EVENTS.ViewBlobPage)
        TELEMETRY_V2_RECORDER.recordEvent('blob', 'view')
    })
</script>

<svelte:head>
    <title>{data.filePath} - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

{#if data.type === 'DiffView'}
    <DiffView {data} />
{:else}
    <FileView bind:this={fileView} {data}>
        <svelte:fragment slot="actions">
            <slot name="actions" />
        </svelte:fragment>
    </FileView>
{/if}
