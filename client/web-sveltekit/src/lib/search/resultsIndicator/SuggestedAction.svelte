<script lang="ts">
    import { capitalizeFirstLetter } from '$lib/search/utils'
    import type { Progress, Skipped } from '$lib/shared'

    export let isError: boolean
    export let progress: Progress
    export let hasSkippedItems: boolean
    export let hasSuggestedItems: boolean
    export let mostSevere: Skipped

    const INTERPUNCT = '\u00B7'

    $: done = progress.done
</script>

{#if progress}
    <div class={`action-container ${isError && 'error-text'}`}>
        <div class="suggested-action">
            {#if done && !hasSkippedItems}
                <div class="more-details">See more details</div>
            {/if}

            {#if done && hasSkippedItems}
                <div class={`info-badge ${isError && 'error-text'}`}>
                    {capitalizeFirstLetter(mostSevere?.title ? mostSevere.title : '')}&nbsp;
                </div>
            {/if}

            {#if done && hasSuggestedItems}
                <div class="separator">{INTERPUNCT}</div>
                <div class="action-badge">
                    {capitalizeFirstLetter(mostSevere?.suggested ? mostSevere.suggested.title : '')}&nbsp;
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
        padding-top: 0.2rem;
        padding-bottom: 0.2rem;
        padding-left: 0.2rem;
        padding-right: 0.2rem;
    }

    .info-badge {
        background-color: var(--primary-2);
        border-radius: 3px;
        color: var(--text-body);
        padding-left: 0.4rem;
        padding-right: 0.2rem;
    }

    .info-badge.duration {
        background: var(--warning-2);
        color: var(--text-body);
    }

    .info-badge.error-text {
        background: var(--danger-2);
        color: var(--text-body);
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
