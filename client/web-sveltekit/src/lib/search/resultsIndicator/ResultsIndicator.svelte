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

    const MAX_SEARCH_DURATION = 15000
    const icons: Record<string, string> = {
        info: mdiInformationOutline,
        warning: mdiAlert,
        error: mdiAlertCircle,
    }

    // @TODO: fix this so that it restarts every time there's a new search.
    onMount(() => {
        const startTime = Date.now()
        setInterval(() => {
            const now = Date.now()
            elapsedDuration = now - startTime
        }, 1300)
    })

    $: elapsedDuration = 0
    $: takingTooLong = elapsedDuration >= MAX_SEARCH_DURATION
    $: mostSevere = sortedItems[0]
    $: isError = state === 'error'
    $: loading = state === 'loading'
    $: severity = progress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error')
        ? 'error'
        : 'info'
    $: done = progress.done
</script>

<div class="indicator">
    <div class="icon">
        {#if loading && !done}
            <LoadingSpinner inline />
        {:else}
            <Icon svgPath={icons[severity]} size={18} />
        {/if}
    </div>

    <div class="messages">
        <ProgressMessage {progress} {loading} {isError} {elapsedDuration} />
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
