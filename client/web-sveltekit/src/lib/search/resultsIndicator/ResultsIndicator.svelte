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

<div class="root">
    <div class="indicator">
        <div class="progress-message">
            <div class="icon">
                {#if loading}
                    <LoadingSpinner --icon-size="18px" inline />
                {:else}
                    <Icon
                        svgPath={icons[severity]}
                        --icon-size="18px"
                        --color={isError ? 'var(--danger)' : 'var(--text-title)'}
                    />
                {/if}
            </div>

            <ProgressMessage {state} {progress} {severity} />
        </div>

        <div class="messages">
            {#if !done && takingTooLong}
                <TimeoutMessage />
            {:else if done}
                <SuggestedAction {progress} {suggestedItems} {severity} {state} />
            {:else}
                <span>Running search...</span>
            {/if}
        </div>
    </div>

    <Icon svgPath={mdiChevronDown} --icon-size="18px" --color={isError ? 'var(--danger)' : 'var(--text-title)'} />
</div>

<style lang="scss">
    .root {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }

    .indicator {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.5rem;

        .progress-message {
            display: flex;
            gap: 0.5rem;
            align-items: center;
        }

        .icon {
            align-self: baseline;
        }

        .messages {
            display: flex;
            flex-flow: row nowrap;
            align-items: baseline;
            gap: 0.25rem;
        }

        span {
            color: var(--text-muted);
        }
    }
</style>
