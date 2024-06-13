<script lang="ts">
    import type { ComponentProps } from 'svelte'

    import Icon2 from '$lib/Icon2.svelte'
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

    const icons: Record<string, ComponentProps<Icon2>['icon']> = {
        info: ILucideInfo,
        warning: ILucideAlertCircle,
        error: ILucideCircleX,
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
    {#if loading}
        <LoadingSpinner --size="16px" />
    {:else}
        <Icon2
            icon={icons[severity]}
            aria-label={severity}
            --icon-size="16px"
            --color={isError ? 'var(--danger)' : 'var(--text-title)'}
        />
    {/if}

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

    <Icon2 icon={ILucideChevronDown} --icon-size="12px" --color={isError ? 'var(--danger)' : 'var(--text-title)'} />
</div>

<style lang="scss">
    .indicator {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: center;
        gap: 0.75rem;
        padding: 0.375rem 0.75rem;
        border-radius: var(--border-radius);

        &:hover {
            background-color: var(--color-bg-2);
        }

        .messages {
            display: flex;
            flex-flow: column nowrap;
            justify-content: center;
            align-items: flex-start;
            row-gap: 0.25rem;
        }

        span {
            color: var(--text-muted);
        }
    }
</style>
