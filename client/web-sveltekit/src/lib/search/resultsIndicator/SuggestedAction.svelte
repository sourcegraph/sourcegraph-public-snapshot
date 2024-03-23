<script lang="ts">
    import { capitalize } from 'lodash'

    import { sortBySeverity } from '$lib/branded'
    import type { Progress } from '$lib/shared'

    export let state: 'loading' | 'error' | 'complete'
    export let progress: Progress
    export let hasSuggestedItems: boolean

    const CENTER_DOT = '\u00B7' // AKA 'interpunct'

    $: sortedItems = sortBySeverity(progress.skipped)
    $: isError = state === 'error'
    $: hasSkippedItems = progress.skipped.length > 0
    $: mostSevere = sortedItems[0]
    $: done = progress.done
</script>

{#if progress}
    <div class="action-container" class:error-text={isError}>
        <div class="suggested-action">
            {#if done && !hasSkippedItems}
                <div class="more-details">See more details</div>
            {/if}

            {#if done && hasSkippedItems}
                <div class="info-badge" class:error-text={isError}>
                    {capitalize(mostSevere?.title ? mostSevere.title : '')}&nbsp;
                </div>
            {/if}

            {#if done && hasSuggestedItems}
                <div class="separator">{CENTER_DOT}</div>
                <div class="action-badge">
                    {capitalize(mostSevere?.suggested ? mostSevere.suggested.title : '')}&nbsp;
                    <span class="code-font">
                        {mostSevere.suggested?.queryExpression}
                    </span>
                </div>
            {/if}
        </div>
    </div>
{/if}

<style lang="scss">
    .action-container {
        margin-top: 0.3rem;
    }

    .code-font {
        background-color: var(--secondary);
        border-radius: 3px;
        color: var(--text-body);
        font-family: var(--code-font-family);
        font-size: 0.7rem;
        padding: 0.2rem;
    }

    .info-badge {
        background-color: var(--primary-2);
        border-radius: 3px;
        color: var(--text-body);
        padding-left: 0.4rem;
        padding-right: 0.2rem;

        &.error-text {
            background: var(--danger-2);
        }
    }

    .more-details {
        color: var(--gray-06);
    }

    .separator {
        margin-left: 0.4rem;
        margin-right: 0.4rem;
    }

    .suggested-action {
        display: flex;
        flex-flow: row nowrap;
        margin-left: 0.2rem;
    }
</style>
