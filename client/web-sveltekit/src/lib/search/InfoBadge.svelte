<script lang="ts">
    import { getProgressText } from '$lib/branded'
    import type { Progress } from '$lib/shared'

    export let searchProgress: Progress
    export let state: 'error' | 'complete' | 'loading'
</script>

<div class={`progress-message ${state === 'error' && 'error-text'}`}>
    {getProgressText(searchProgress).visibleText}
</div>

<style lang="scss">
    .progress-message {
        font-size: 0.9rem;
    }

    .error-text {
        color: var(--danger);
    }
</style>
