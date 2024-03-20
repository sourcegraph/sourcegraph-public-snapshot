<script lang="ts">
    import { capitalizeFirstLetter } from '$lib/search/utils'
    import type { Progress, Skipped } from '$lib/shared'

    export let isError: boolean
    export let loading: boolean
    export let progress: Progress
    export let isComplete: boolean
    export let hasSkippedItems: boolean
    export let hasSuggestedItems: boolean
    export let mostSevere: Skipped

    const CENTER_DOT = '\u00B7' // interpunct

    $: isComplete = isComplete
    $: isError = isError
    $: loading = loading
    $: progress = progress
</script>

{#if !loading && progress}
    <div class={`action-container ${isError && 'error-text'}`}>
        <div class="suggested-action">
            {#if !loading && isComplete && !hasSkippedItems}
                <div class="more-details">See more details</div>
            {/if}

            {#if mostSevere}
                <div class="info-badge">
                    {capitalizeFirstLetter(mostSevere?.title ? mostSevere.title : '')}&nbsp;
                </div>
            {/if}

            {#if hasSkippedItems && hasSuggestedItems}
                <div class="separator">{CENTER_DOT}</div>
            {/if}

            {#if hasSuggestedItems}
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
        background-color: var(--gray-06);
        border-radius: 3px;
        color: white;
        font-family: var(--code-font-family);
        font-size: 0.8rem;
        padding-left: 0.2rem;
        padding-right: 0.2rem;
    }

    .info-badge {
        background-color: var(--primary);
        border-radius: 3px;
        color: white;
        padding-left: 0.2rem;
        padding-right: 0.2rem;
    }

    .info-badge.duration {
        background: var(--warning);
        color: black;
    }

    .info-badge.error {
        background: var(--danger);
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
