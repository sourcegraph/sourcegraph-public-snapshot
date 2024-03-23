<script lang="ts">
    import InfoBadge from '$lib/search/resultsIndicator/InfoBadge.svelte'
    import type { Progress } from '$lib/shared'

    export let state: 'error' | 'loading' | 'complete'
    export let progress: Progress
    export let elapsedDuration: number
    export let threshold: number

    $: isError = state === 'error'
    $: loading = state === 'loading'
</script>

{#if loading}
    <div class="progress-message">
        Fetching results... {(elapsedDuration / 1000).toFixed(1)}s
    </div>
    <div class="action-container" class:error-text={isError}>
        <div class="suggested-action">
            {#if elapsedDuration <= threshold}
                <div class="running-search">
                    <small> Running Search </small>
                </div>
            {/if}
        </div>
    </div>
{:else}
    <InfoBadge {state} {progress} />
{/if}

<style lang="scss">
    .action-container {
        margin-top: 0.3rem;
    }

    .error-text {
        color: var(--danger);
    }

    .progress-message {
        margin-left: 0.2rem;
        font-size: var(--font-size-base);
    }

    .suggested-action {
        display: flex;
        flex-flow: row nowrap;
        margin-left: 0.2rem;
    }

    .running-search {
        color: var(--text-muted);
    }
</style>
