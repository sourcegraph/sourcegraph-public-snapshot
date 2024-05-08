<svelte:options immutable />

<script lang="ts">
    // @sg EnableRollout
    import { onMount } from 'svelte'

    import { SVELTE_LOGGER, SVELTE_TELEMETRY_EVENTS } from '$lib/telemetry'

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

    let fileView: FileView

    onMount(() => {
        SVELTE_LOGGER.logViewEvent(SVELTE_TELEMETRY_EVENTS.ViewBlobPage)
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
