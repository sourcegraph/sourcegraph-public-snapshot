<script lang="ts">
    import { getProgressText } from '$lib/branded'
    import type { Progress } from '$lib/shared'

    export let state: 'error' | 'loading' | 'complete'
    export let progress: Progress
    export let severity: string

    $: isError = state === 'error' || severity === 'error'
    $: loading = state === 'loading'
</script>

<div class="progress-message" class:error-text={isError}>
    {#if loading}
        Fetching results...
    {:else}
        {getProgressText(progress).visibleText}
    {/if}
</div>

<style lang="scss">
    .progress-message {
        font-size: var(--font-size-small);
        color: var(--text-title);

        &.error-text {
            color: var(--danger);
        }
    }
</style>
