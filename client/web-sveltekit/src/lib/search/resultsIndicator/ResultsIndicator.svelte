<script lang="ts">
    import { mdiChevronDown, mdiInformationOutline, mdiAlert, mdiAlertCircle } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import ProgressMessage from '$lib/search/resultsIndicator/ProgressMessage.svelte'
    import SuggestedAction from '$lib/search/resultsIndicator/SuggestedAction.svelte'
    import TimeoutMessage from '$lib/search/resultsIndicator/TimeoutMessage.svelte'
    import type { Progress, Skipped } from '$lib/shared'

    export let state: 'error' | 'complete' | 'loading'
    export let progress: Progress
    export let suggestedItems: Required<Skipped>[]
    export let severity: string

    const SEARCH_JOB_THRESHOLD = 10000

    const icons: Record<string, string> = {
        info: mdiInformationOutline,
        warning: mdiAlert,
        error: mdiAlertCircle,
    }

    $: elapsedDuration = progress.durationMs
    $: takingTooLong = elapsedDuration >= SEARCH_JOB_THRESHOLD
    $: loading = state === 'loading'
    $: severity = progress.skipped.some(skipped => skipped.severity === 'warn' || skipped.severity === 'error')
        ? 'error'
        : 'info'
    $: isError = severity === 'error' || state === 'error'
    /*
     * NOTE: progress.done and state === 'complete' will sometimes be different values.
     * Only one of them needs to evaluate to true in order for the ResultIndicator to
     * evaluate a search as being finished. Hence, we check both here with an OR relationship
     */
    $: done = progress.done || state === 'complete'
</script>

<div class="indicator">
    <div>
        {#if loading}
            <LoadingSpinner inline />
        {:else}
            <Icon svgPath={icons[severity]} size={18} --color={isError ? 'var(--danger)' : 'var(--text-title)'} />
        {/if}
    </div>

    <div class="messages">
        <ProgressMessage {state} {progress} {severity} />
        {#if !done && takingTooLong}
            <TimeoutMessage />
        {:else if done}
            <SuggestedAction {progress} {suggestedItems} {severity} {state} />
        {:else}
            <span>Running search...</span>
        {/if}
    </div>
    <Icon svgPath={mdiChevronDown} size={18} --color={isError ? 'var(--danger)' : 'var(--text-title)'} />
</div>

<style lang="scss">
    .indicator {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-evenly;
        align-items: center;
        gap: 0.5rem;
        min-width: 200px;
        max-width: fit-content;

        padding: 0.25rem;

        .messages {
            display: flex;
            flex-flow: column nowrap;
            justify-content: center;
            align-items: flex-start;
            margin-right: 0.75rem;
            margin-left: 0.5rem;
            row-gap: 0.25rem;
        }

        span {
            color: var(--text-muted);
        }
    }
</style>
