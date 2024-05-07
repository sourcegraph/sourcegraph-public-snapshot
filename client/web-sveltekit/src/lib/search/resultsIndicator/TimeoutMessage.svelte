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
        column-gap: 0.5rem;
        color: var(--text-muted);
        white-space: nowrap;

        .info-badge {
            background-color: var(--warning-2);
            border-radius: 3px;
            padding: 0rem 0.15rem;
            color: var(--text-title);
            white-space: nowrap;
        }
    }
</style>
