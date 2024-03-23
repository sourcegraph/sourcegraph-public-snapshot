<script lang="ts">
    import { getProgressText } from '$lib/branded'
    import type { Progress } from '$lib/shared'

    export let searchProgress: Progress
    export let state: 'error' | 'complete' | 'loading'

    $: isError = state === 'error'
</script>

<div class="progress-message" class:error-text={isError}>
    {getProgressText(searchProgress).visibleText}
</div>

<style lang="scss">
    .error-text {
        color: var(--danger);
    }

    .progress-message {
        font-size: 0.9rem;
        margin-left: 0.2rem;
    }
</style>
