<script lang="ts">
    import { mdiChevronDown, mdiInformationOutline, mdiAlert, mdiAlertCircle } from '@mdi/js'
    import { onMount } from 'svelte'

    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import ProgressMessage from '$lib/search/resultsIndicator/ProgressMessage.svelte'
    import SuggestedAction from '$lib/search/resultsIndicator/SuggestedAction.svelte'
    import TimeoutMessage from '$lib/search/resultsIndicator/TimeoutMessage.svelte'
    import type { Progress, Skipped } from '$lib/shared'

    export let hasSkippedItems: boolean
    export let sortedItems: Skipped[]
    export let hasSuggestedItems: boolean
    export let progress: Progress
    export let state: 'error' | 'complete' | 'loading'

    const SEARCH_JOB_THRESHOLD = 8000
    const icons: Record<string, string> = {
        info: mdiInformationOutline,
        warning: mdiAlert,
        error: mdiAlertCircle,
    }

    onMount(() => {
        let startTime = Date.now()
        setInterval(() => {
            const now = Date.now()
            elapsedDuration = now - startTime
            // once search has completed, reset the startTime
            if (done) {
                startTime = Date.now()
            }
        }, 1300)
    })

    $: elapsedDuration = 0
    $: takingTooLong = elapsedDuration >= SEARCH_JOB_THRESHOLD
    $: mostSevere = sortedItems[0]
    $: isError = state === 'error'
    $: loading = state === 'loading'
    $: severity = progress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error')
        ? 'error'
        : 'info'
    /*
     * TODO: @jasonhawkharris Explore combining 'complete' and 'done' values.
     * The values do refer to different objects. 'complete' is the state returned
     * from the stream. 'done' is the state returned from the progress object (which)
     * also exists as a smaller object inside of the search stream. Do they need to?
     */
    $: done = progress.done || state === 'complete'
</script>

<div class="indicator">
    <div class="icon">
        {#if loading && !hasSkippedItems}
            <LoadingSpinner inline />
        {:else}
            <Icon svgPath={icons[severity]} size={18} />
        {/if}
    </div>

    <div class="messages">
        <ProgressMessage
            {state}
            {progress}
            {loading}
            {isError}
            {elapsedDuration}
            searchJobThreshold={SEARCH_JOB_THRESHOLD}
        />
        {#if !done && takingTooLong}
            <TimeoutMessage {isError} />
        {/if}
        <SuggestedAction {isError} {progress} {hasSkippedItems} {hasSuggestedItems} {mostSevere} />
    </div>

    <div class="dropdown-icon">
        <Icon svgPath={mdiChevronDown} size={18} />
    </div>
</div>

<style lang="scss">
    .dropdown-icon {
        margin-left: 2rem;
    }

    .indicator {
        align-items: center;
        display: flex;
        flex-flow: row nowrap;
    }

    .messages {
        align-content: center;
        align-items: flex-start;
        display: flex;
        flex-flow: column nowrap;
        margin-left: 0.5rem;
    }
</style>
