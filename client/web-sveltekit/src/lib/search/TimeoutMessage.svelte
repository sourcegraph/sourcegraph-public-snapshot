<script lang="ts">
    export let isError: boolean
    export let loading: boolean
    export let isComplete: boolean
    export let takingTooLong: boolean

    const CENTER_DOT = '\u00B7' // interpunct
</script>

{#if loading && takingTooLong && !isComplete}
    <div class={`info-badge ${isError && 'error-text'}`}>
        <div class="duration-badge">
            <div class="info-badge duration">Taking too long?</div>
            <div class="separator">{CENTER_DOT}</div>
            <div class="action-badge">
                Use
                <a href="https://sourcegraph.com/docs/code-search/types/search-jobs" target="_blank"> Search Job </a>
                for background search
            </div>
        </div>
    </div>
{:else if loading && !takingTooLong}
    {#if isComplete}
        <div class="more-details">See more details</div>
    {:else}
        <div class="loading-action-message">Running search...</div>
    {/if}
{/if}

<style lang="scss">
    .duration-badge {
        display: flex;
        flex-flow: row-nowrap;
        margin-top: 0.2rem;
    }

    .info-badge {
        color: var(--gray-08);
        border-radius: 3px;
        padding-left: 0.2rem;
        padding-right: 0.2rem;
    }

    .loading-action-message {
        margin-left: 0.2rem;
    }

    .more-details {
        color: var(--gray-06);
        margin-left: 0.2rem;
    }

    .error-text {
        color: var(--danger);
    }

    .info-badge.duration {
        background: var(--warning);
        color: black;
    }

    .progress-message {
        font-size: 0.9rem;
    }

    .error-text {
        color: var(--danger);
    }

    .separator {
        margin-right: 0.4rem;
        margin-left: 0.4rem;
    }
</style>
