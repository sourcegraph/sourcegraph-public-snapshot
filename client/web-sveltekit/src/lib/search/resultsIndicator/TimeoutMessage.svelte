<script lang="ts">
    export let isError: boolean

    const CENTER_DOT = '\u00B7' // interpunct
</script>

<div class="action-container" class:error-text={isError}>
    <div class="suggested-action">
        <div class="info-badge">Taking too long?</div>
        <div class="separator">{CENTER_DOT}</div>
        <div class="action-badge">
            Use
            <a href="https://sourcegraph.com/docs/code-search/types/search-jobs" target="_blank"> Search Job </a>
            for background search
        </div>
    </div>
</div>

<style lang="scss">
    .action-container {
        margin-top: 0.3rem;
    }

    .info-badge {
        background-color: var(--warning-2);
        border-radius: 3px;
        color: var(--text-body);
        padding-left: 0.4rem;
        padding-right: 0.2rem;
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
