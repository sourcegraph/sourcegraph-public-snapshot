<script lang="ts">
    export let isError: boolean

    const CENTER_DOT = '\u00B7' // interpunct
</script>

<div class="info-badge" class:error-text={isError}>
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

<style lang="scss">
    .duration-badge {
        display: flex;
        flex-flow: row-nowrap;
    }

    .error-text {
        color: var(--danger);
    }

    .info-badge {
        border-radius: 3px;
        color: var(--text-body);
        padding: 0 0.2rem;

        &.duration {
            background: var(--warning-2);
        }
    }

    .separator {
        margin-left: 0.4rem;
        margin-right: 0.4rem;
    }
</style>
