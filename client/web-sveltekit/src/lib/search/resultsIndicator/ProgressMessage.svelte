<script lang="ts">
    import InfoBadge from '$lib/search/resultsIndicator/InfoBadge.svelte'
    import type { Progress } from '$lib/shared'

    export let state: 'error' | 'loading' | 'complete'
    export let progress: Progress
    export let loading: boolean
    export let isError: boolean
    export let elapsedDuration: number
    export let maxSearchDuration: number
</script>

{#if loading}
    <div class="progress-message">Fetching results... {(elapsedDuration / 1000).toFixed(1)}s</div>
    <div class={`action-container ${isError && 'error-text'}`}>
        <div class="suggested-action">
            {#if elapsedDuration <= maxSearchDuration}
                <div class="running-search">Running Search</div>
            {/if}
        </div>
    </div>
{:else}
    <InfoBadge {state} searchProgress={progress} />
{/if}

<style lang="scss">
    .action-container {
        margin-top: 0.3rem;
    }

    .error-text {
        color: var(--danger);
    }

    .progress-message {
        font-size: 0.9rem;
        margin-left: 0.2rem;
    }

    .suggested-action {
        display: flex;
        flex-flow: row nowrap;
        margin-left: 0.2rem;
    }

    .running-search {
        color: var(--gray-06);
    }
</style>
