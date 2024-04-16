<script lang="ts">
    const CENTER_DOT = '\u00B7' // interpunct
</script>

<div>
    <small class="info-badge">Taking too long?</small>
    <small>{CENTER_DOT}</small>
    <small>
        Use
        <a href="/help/code-search/types/search-jobs" target="_blank" rel="noopener noreferrer"> Search Job </a>
        for background search
    </small>
</div>

<style lang="scss">
    div {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        justify-content: flex-end;
        gap: 0.5rem 0.5rem;

        .info-badge {
            background-color: var(--warning-2);
            border-radius: 3px;
            padding-left: 0.25rem;
            padding-right: 0.25rem;
        }
    }
</style>
