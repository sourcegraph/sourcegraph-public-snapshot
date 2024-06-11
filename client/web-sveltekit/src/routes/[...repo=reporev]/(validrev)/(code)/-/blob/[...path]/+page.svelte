<svelte:options immutable />

<script lang="ts">
    // @sg EnableRollout
    import { onMount } from 'svelte'

    import { TELEMETRY_V2_RECORDER } from '$lib/telemetry2'

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
