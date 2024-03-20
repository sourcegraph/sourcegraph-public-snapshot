<script lang="ts">
    import { getProgressText } from '$lib/branded'
    import type { Progress } from '$lib/shared'

    export let searchProgress: Progress
    export let loading: boolean
    export let isError: boolean
    export let elapsedDuration: number
</script>

{#if !loading}
    <div class="progress-message">Fetching results... {(elapsedDuration / 1000).toFixed(1)}s</div>
{:else}
    <div class={`progress-message ${isError && 'error-text'}`}>
        {getProgressText(searchProgress).visibleText}
    </div>
{/if}

<style lang="scss">
    .error-text {
        color: var(--danger);
    }

    .progress-message {
        font-size: 0.9rem;
        margin-left: 0.2rem;
    }
</style>
